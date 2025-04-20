package service_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansoncht/flight-microservices/internal/reader/client"
	"github.com/ansoncht/flight-microservices/internal/reader/model"
	"github.com/ansoncht/flight-microservices/internal/reader/service"
	"github.com/ansoncht/flight-microservices/internal/reader/service/mocks"
	"github.com/ansoncht/flight-microservices/pkg/kafka"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewReader_NonNilClients_ShouldSucceed(t *testing.T) {
	t.Run("Valid Config", func(t *testing.T) {
		actual, err := service.NewReader(&client.FlightAPIClient{}, &client.RouteAPIClient{}, &kafka.Writer{})

		require.NoError(t, err)
		require.NotNil(t, actual)
	})
}

func TestNewReader_NilClient_ShouldError(t *testing.T) {
	tests := []struct {
		name          string
		flightClient  client.FlightsClient
		routeClient   client.RoutesClient
		messageWriter kafka.MessageWriter
		wantErr       string
	}{
		{
			name:          "Nil Flight API Client",
			flightClient:  nil,
			routeClient:   &client.RouteAPIClient{},
			messageWriter: &kafka.Writer{},
			wantErr:       "flight client is nil",
		},
		{
			name:          "Nil Route API Client",
			flightClient:  &client.FlightAPIClient{},
			routeClient:   nil,
			messageWriter: &kafka.Writer{},
			wantErr:       "route client is nil",
		},
		{
			name:          "Nil Kafka Writer",
			flightClient:  &client.FlightAPIClient{},
			routeClient:   &client.RouteAPIClient{},
			messageWriter: nil,
			wantErr:       "message writer is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := service.NewReader(tt.flightClient, tt.routeClient, tt.messageWriter)

			require.Nil(t, actual)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestHTTPHandler_MissingAirport_ShouldError(t *testing.T) {
	r, err := service.NewReader(&client.FlightAPIClient{}, &client.RouteAPIClient{}, &kafka.Writer{})

	require.NoError(t, err)
	require.NotNil(t, r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fetch", nil)
	w := httptest.NewRecorder()

	r.HTTPHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "missing airport parameter\n", w.Body.String())
}

func TestHTTPHandler_WorkingComponents_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mRoutes := mocks.NewMockRoutesClient(ctrl)
	mFlights := mocks.NewMockFlightsClient(ctrl)
	mKafka := mocks.NewMockMessageWriter(ctrl)

	flights := []model.Flight{
		{Origin: "VHHH", Destination: "RJTT", Callsign: "CRK452", FirstSeen: 1, LastSeen: 2},
	}
	route := &model.Route{
		Response: model.Response{
			FlightRoute: model.FlightRoute{
				CallSignIATA: "CRK452",
				Airline:      model.Airline{Name: "PAL"},
				Origin:       model.Airport{IATACode: "VHHH"},
				Destination:  model.Airport{IATACode: "RJTT"},
			},
		},
	}

	mFlights.EXPECT().FetchFlights(ctx, "VHHH", gomock.Any(), gomock.Any()).Return(flights, nil)
	mRoutes.EXPECT().FetchRoute(gomock.Any(), "CRK452").Return(route, nil)
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fetch?airport=VHHH", nil)
	w := httptest.NewRecorder()

	r, err := service.NewReader(mFlights, mRoutes, mKafka)

	require.NoError(t, err)
	require.NotNil(t, r)

	r.HTTPHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, w.Body.String(), "Flights processed successfully")
}

func TestHTTPHandler_FlightsClientError_ShouldError(t *testing.T) {
	ctrl := gomock.NewController(t)

	m := mocks.NewMockFlightsClient(ctrl)

	m.EXPECT().FetchFlights(context.Background(), "VHHH", gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	r, err := service.NewReader(m, &client.RouteAPIClient{}, &kafka.Writer{})

	require.NoError(t, err)
	require.NotNil(t, r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fetch?airport=VHHH", nil)
	w := httptest.NewRecorder()

	r.HTTPHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Contains(t, w.Body.String(), "failed to process flights")
}

func TestHTTPHandler_RoutesClientError_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mRoutes := mocks.NewMockRoutesClient(ctrl)
	mFlights := mocks.NewMockFlightsClient(ctrl)

	flights := []model.Flight{
		{Origin: "VHHH", Destination: "RJTT", Callsign: "CRK452", FirstSeen: 1, LastSeen: 2},
	}

	mFlights.EXPECT().FetchFlights(ctx, "VHHH", gomock.Any(), gomock.Any()).Return(flights, nil)
	mRoutes.EXPECT().FetchRoute(gomock.Any(), "CRK452").Return(nil, errors.New("error"))

	r, err := service.NewReader(mFlights, mRoutes, &kafka.Writer{})

	require.NoError(t, err)
	require.NotNil(t, r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fetch?airport=VHHH", nil)
	w := httptest.NewRecorder()

	r.HTTPHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, w.Body.String(), "Flights processed successfully")
}

func TestHTTPHandler_MessageWriterError_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mRoutes := mocks.NewMockRoutesClient(ctrl)
	mFlights := mocks.NewMockFlightsClient(ctrl)
	mKafka := mocks.NewMockMessageWriter(ctrl)

	flights := []model.Flight{
		{Origin: "VHHH", Destination: "RJTT", Callsign: "CRK452", FirstSeen: 1, LastSeen: 2},
	}

	route := &model.Route{
		Response: model.Response{
			FlightRoute: model.FlightRoute{
				CallSignIATA: "CRK452",
				Airline:      model.Airline{Name: "PAL"},
				Origin:       model.Airport{IATACode: "VHHH"},
				Destination:  model.Airport{IATACode: "RJTT"},
			},
		},
	}

	mFlights.EXPECT().FetchFlights(ctx, "VHHH", gomock.Any(), gomock.Any()).Return(flights, nil)
	mRoutes.EXPECT().FetchRoute(gomock.Any(), "CRK452").Return(route, nil)
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("error"))

	r, err := service.NewReader(mFlights, mRoutes, mKafka)

	require.NoError(t, err)
	require.NotNil(t, r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fetch?airport=VHHH", nil)
	w := httptest.NewRecorder()

	r.HTTPHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, w.Body.String(), "Flights processed successfully")
}
