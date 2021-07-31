LDFLAGS = -X 'sci_hub_p2p/pkg/vars.Ref=${REF}'
LDFLAGS += -X 'sci_hub_p2p/pkg/vars.Commit=${SHA}'
LDFLAGS += -X 'sci_hub_p2p/pkg/vars.Builder=$(shell go version)'
LDFLAGS += -X 'sci_hub_p2p/pkg/vars.BuildTime=${TIME}'

GoBuildArgs = -ldflags "-s -w $(LDFLAGS)" -tags disable_libutp
