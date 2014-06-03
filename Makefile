# usage : make PREFIX=$prefix

SRC=$(filter-out %_test.go,$(filter-out ./bin/%,$(shell find . -name "*.go")))
CONF=$(shell find etc -name "*.conf")
BIN=$(basename $(wildcard bin/*.go))
GXX=go build
LDPATH=GOPATH=$(GOPATH)

PREFIX?=/usr/local/goatherd
PREFIX_BIN=$(PREFIX)/bin
PREFIX_CONF=$(PREFIX)/etc

all: $(BIN)

$(BIN) : % : %.go $(SRC)
	$(LDPATH) $(GXX) -o $@ $<

install:
	@test -d $(PREFIX_BIN) || mkdir -p $(PREFIX_BIN)
	@for bin in $(BIN);do	\
	    install $${bin} $(PREFIX_BIN)/`basename $${bin}`; \
	done
	@test -d $(PREFIX_CONF) || mkdir -p $(PREFIX_CONF)
	@for conf in $(CONF);do	\
	    test -f $(PREFIX)/$${conf} || cp $${conf} $(PREFIX_CONF)/; \
	done

clean:
	rm -f $(GOROOT)/bin/goatherd*
	rm -f $(BIN)
