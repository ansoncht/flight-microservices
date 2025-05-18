.PHONY: help
help:
	@echo "Available make targets:"
	@grep '^\.PHONY:' Makefile | sed 's/\.PHONY: //g' | tr ' ' '\n' | sort | uniq | \
	while read target; do \
		echo "  make $$target"; \
	done

.PHONY: reader
reader:
	docker build -t flight-reader -f docker/reader.Dockerfile .

.PHONY: processor
processor:
	docker build -t flight-processor -f docker/processor.Dockerfile .

.PHONY: poster
poster:
	docker build -t flight-poster -f docker/poster.Dockerfile .
