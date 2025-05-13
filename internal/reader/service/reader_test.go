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
	"github.com/ansoncht/flight-microservices/internal/test/mock"
	"github.com/ansoncht/flight-microservices/pkg/kafka"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewReader_NonNilClients_ShouldSucceed(t *testing.T) {
	reader, err := service.NewReader(&client.FlightAPI{}, &client.RouteAPI{}, &kafka.Writer{})
	require.NoError(t, err)
	require.NotNil(t, reader)
}

func TestNewReader_NilClient_ShouldError(t *testing.T) {
	tests := []struct {
		name          string
		flightClient  client.Flight
		routeClient   client.Route
		messageWriter kafka.MessageWriter
		wantErr       string
	}{
		{
			name:          "Nil Flight Client",
			flightClient:  nil,
			routeClient:   &client.RouteAPI{},
			messageWriter: &kafka.Writer{},
			wantErr:       "flight client is nil",
		},
		{
			name:          "Nil Route Client",
			flightClient:  &client.FlightAPI{},
			routeClient:   nil,
			messageWriter: &kafka.Writer{},
			wantErr:       "route client is nil",
		},
		{

			name:          "Nil Message Writer",
			flightClient:  &client.FlightAPI{},
			routeClient:   &client.RouteAPI{},
			messageWriter: nil,
			wantErr:       "message writer is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := service.NewReader(tt.flightClient, tt.routeClient, tt.messageWriter)
			require.Nil(t, reader)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestHTTPHandler_MissingAirport_ShouldError(t *testing.T) {
	reader, err := service.NewReader(&client.FlightAPI{}, &client.RouteAPI{}, &kafka.Writer{})
	require.NoError(t, err)
	require.NotNil(t, reader)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fetch", nil)
	w := httptest.NewRecorder()
	reader.HTTPHandler(w, req)

	resp := w.Result()
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "missing airport parameter\n", w.Body.String())
}

func TestHTTPHandler_WorkingComponents_ShouldSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mRoutes := mock.NewMockRoute(ctrl)
	mFlights := mock.NewMockFlight(ctrl)
	mKafka := mock.NewMockMessageWriter(ctrl)

	flights := []model.Flight{
		{Origin: "VHHH", Destination: "RJTT", Callsign: "CRK452", FirstSeen: 1, LastSeen: 2},
	}

	mFlights.EXPECT().FetchFlights(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(flights, nil)
	mRoutes.EXPECT().FetchRoute(gomock.Any(), gomock.Any()).Return(&model.Route{}, nil)
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	reader, err := service.NewReader(mFlights, mRoutes, mKafka)
	require.NoError(t, err)
	require.NotNil(t, reader)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fetch?airport=VHHH", nil)
	w := httptest.NewRecorder()
	reader.HTTPHandler(w, req)

	resp := w.Result()
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, w.Body.String(), "flights processed successfully")
}

func TestHTTPHandler_FlightsClientError_ShouldError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mFlights := mock.NewMockFlight(ctrl)

	mFlights.EXPECT().FetchFlights(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	reader, err := service.NewReader(mFlights, &client.RouteAPI{}, &kafka.Writer{})
	require.NoError(t, err)
	require.NotNil(t, reader)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fetch?airport=VHHH", nil)
	w := httptest.NewRecorder()
	reader.HTTPHandler(w, req)

	resp := w.Result()
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Contains(t, w.Body.String(), "failed to process flights")
}

func TestHTTPHandler_EmptyCallSign_ShouldSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mRoutes := mock.NewMockRoute(ctrl)
	mFlights := mock.NewMockFlight(ctrl)
	mKafka := mock.NewMockMessageWriter(ctrl)

	flights := []model.Flight{
		{Origin: "VHHH", Destination: "RJTT", Callsign: "", FirstSeen: 1, LastSeen: 2},
	}

	mFlights.EXPECT().FetchFlights(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(flights, nil)
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	reader, err := service.NewReader(mFlights, mRoutes, mKafka)
	require.NoError(t, err)
	require.NotNil(t, reader)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fetch?airport=VHHH", nil)
	w := httptest.NewRecorder()
	reader.HTTPHandler(w, req)

	resp := w.Result()
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, w.Body.String(), "flights processed successfully")
}

func TestHTTPHandler_RoutesClientError_ShouldSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mRoutes := mock.NewMockRoute(ctrl)
	mFlights := mock.NewMockFlight(ctrl)
	mKafka := mock.NewMockMessageWriter(ctrl)

	flights := []model.Flight{
		{Origin: "VHHH", Destination: "RJTT", Callsign: "CRK452", FirstSeen: 1, LastSeen: 2},
	}

	mFlights.EXPECT().FetchFlights(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(flights, nil)
	mRoutes.EXPECT().FetchRoute(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	reader, err := service.NewReader(mFlights, mRoutes, mKafka)
	require.NoError(t, err)
	require.NotNil(t, reader)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fetch?airport=VHHH", nil)
	w := httptest.NewRecorder()
	reader.HTTPHandler(w, req)

	resp := w.Result()
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, w.Body.String(), "flights processed successfully")
}

func TestHTTPHandler_MessageWriterError_ShouldSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mRoutes := mock.NewMockRoute(ctrl)
	mFlights := mock.NewMockFlight(ctrl)
	mKafka := mock.NewMockMessageWriter(ctrl)

	flights := []model.Flight{
		{Origin: "VHHH", Destination: "RJTT", Callsign: "CRK452", FirstSeen: 1, LastSeen: 2},
	}

	mFlights.EXPECT().FetchFlights(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(flights, nil)
	mRoutes.EXPECT().FetchRoute(gomock.Any(), gomock.Any()).Return(&model.Route{}, nil)
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("error"))
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	reader, err := service.NewReader(mFlights, mRoutes, mKafka)
	require.NoError(t, err)
	require.NotNil(t, reader)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fetch?airport=VHHH", nil)
	w := httptest.NewRecorder()
	reader.HTTPHandler(w, req)

	resp := w.Result()
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, w.Body.String(), "flights processed successfully")
}

func TestHTTPHandler_RouteProcessContextCancellation_ShouldError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mRoutes := mock.NewMockRoute(ctrl)
	mFlights := mock.NewMockFlight(ctrl)
	mKafka := mock.NewMockMessageWriter(ctrl)

	flights := []model.Flight{
		{Origin: "VHHH", Destination: "RJTT", Callsign: "CRK452", FirstSeen: 1, LastSeen: 2},
	}

	mFlights.EXPECT().FetchFlights(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(flights, nil)
	mRoutes.EXPECT().FetchRoute(gomock.Any(), gomock.Any()).Return(nil, context.Canceled)
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	reader, err := service.NewReader(mFlights, mRoutes, mKafka)
	require.NoError(t, err)
	require.NotNil(t, reader)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fetch?airport=VHHH", nil)
	w := httptest.NewRecorder()
	reader.HTTPHandler(w, req)

	resp := w.Result()
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Contains(t, w.Body.String(), "failed to process at least one route")
}

func TestHTTPHandler_MessageProcessContextCancellation_ShouldError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mRoutes := mock.NewMockRoute(ctrl)
	mFlights := mock.NewMockFlight(ctrl)
	mKafka := mock.NewMockMessageWriter(ctrl)

	flights := []model.Flight{
		{Origin: "VHHH", Destination: "RJTT", Callsign: "CRK452", FirstSeen: 1, LastSeen: 2},
	}

	mFlights.EXPECT().FetchFlights(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(flights, nil)
	mRoutes.EXPECT().FetchRoute(gomock.Any(), gomock.Any()).Return(&model.Route{}, nil)
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mKafka.EXPECT().WriteMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(context.Canceled)

	reader, err := service.NewReader(mFlights, mRoutes, mKafka)
	require.NoError(t, err)
	require.NotNil(t, reader)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fetch?airport=VHHH", nil)
	w := httptest.NewRecorder()
	reader.HTTPHandler(w, req)

	resp := w.Result()
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Contains(t, w.Body.String(), "failed to process at least one route")
}

func TestClose_ValidAction_ShouldSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mKafka := mock.NewMockMessageWriter(ctrl)

	mKafka.EXPECT().Close().Return()

	reader, err := service.NewReader(&client.FlightAPI{}, &client.RouteAPI{}, mKafka)
	require.NoError(t, err)
	require.NotNil(t, reader)
	defer reader.Close()
}
