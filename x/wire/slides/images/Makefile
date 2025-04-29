a2ss = $(wildcard *.a2s)
a2ssvgs = $(a2ss:.a2s=.svg)

dots = $(wildcard *.dot)
dotsvgs = $(dots:.dot=.svg)

all: $(a2ssvgs) $(dotsvgs)

%.svg: %.a2s
	a2s < $< > $@

%.svg: %.dot
	dot -Tsvg $< > $@
