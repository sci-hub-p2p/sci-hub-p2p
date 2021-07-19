LDFLAGS = -X 'sci_hub_p2p/pkg/variable.Ref=${REF}'
LDFLAGS += -X 'sci_hub_p2p/pkg/variable.Commit=${SHA}'
LDFLAGS += -X 'sci_hub_p2p/pkg/variable.Builder=$(shell go version)'
LDFLAGS += -X 'sci_hub_p2p/pkg/variable.BuildTime=${TIME}'

GoBuildArgs = -ldflags "-s -w $(LDFLAGS)" -tags disable_libutp
