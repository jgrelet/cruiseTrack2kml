VERSION = 0.3.3
BINARY = cruiseTrack2kml

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.Version=${VERSION}  \
-X main.BuildTime=`TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ'`"

#PLATFORMS := linux/amd64 windows/amd64 linux/arm darwin/amd64
PLATFORMS := linux/amd64 windows/amd64 

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

all:
	go build -$(LDFLAGS)

release: $(PLATFORMS)

$(PLATFORMS):
ifeq ($(os),linux)	
	GOOS=$(os) GOARCH=$(arch) go build -o $(BINARY)-'$(os)-$(arch)'.exe -$(LDFLAGS)	
else
	GOOS=$(os) GOARCH=$(arch) go build -o $(BINARY)-'$(os)-$(arch)' -$(LDFLAGS)
endif

clean:
	-rm -f $(BINARY)-*
	-rm -f $(BINARY).exe
	
.PHONY: release $(PLATFORMS) clean run allos simulgps