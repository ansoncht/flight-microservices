PROTO_DIR = proto
OUTPUT_DIR = src

.PHONY: summarizer
summarizer:
	@mkdir -p ${PROTO_DIR}/${OUTPUT_DIR}/summarizer
	@protoc \
	-I${PROTO_DIR} \
	--go_out=${PROTO_DIR}/${OUTPUT_DIR}/summarizer --go_opt=paths=source_relative \
	--go-grpc_out=${PROTO_DIR}/${OUTPUT_DIR}/summarizer --go-grpc_opt=paths=source_relative \
	${PROTO_DIR}/summarizer.proto
