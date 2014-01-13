SRC=$(shell find . -name "*.go")
BIN=$(basename $(wildcard bin/*/*.go))
GXX=go

all : $(BIN)
$(BIN) : % : %.go $(SRC)
	$(GXX) build -o $(@D)/goatherd_$(notdir $(@D)) $<

update:
	@make clean
	@make

clean:
	@rm -f $(BIN)
