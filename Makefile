release-test: SHELL:=/bin/bash
release-test:
	./scripts/run-bosh-release-tests.sh

unit-test:
	bundle exec rspec spec
