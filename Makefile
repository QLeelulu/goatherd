# usage : make -PREFIX=/usr/local/goatherd

SRC=$(filter-out ./bin/%,$(shell find . -name "*.go"))
CONF=$(shell find etc -name "*.conf")
BIN=$(addprefix .bin/gotherd_,$(notdir $(wildcard bin/*)))
GXX=go build

PREFIX=/usr/local/goatherd
PREFIX_BIN=$(PREFIX)/bin
PREFIX_CONF=$(PREFIX)/etc

all : $(BIN)
$(BIN) : .bin/gotherd_% : bin/%/main.go $(SRC)
	$(GXX) -o $@ $<

install:
	@test -d $(PREFIX_BIN) || mkdir -p $(PREFIX_BIN)
	@test -d $(PREFIX_CONF) || mkdir -p $(PREFIX_CONF)
	@for bin in $(BIN);do	\
	    echo install `basename $${bin}`;\
	    install $${bin} $(PREFIX_BIN)/`basename $${bin}`; \
	done
	@for conf in $(CONF);do	\
	    echo copy $${conf}; \
	    test -f $(PREFIX_BIN) || cp $${conf} $(PREFIX_CONF)/; \
	done

clean:
	@rm -f $(BIN)
