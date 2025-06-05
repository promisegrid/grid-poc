#!/bin/bash
# Run simulation nodes in separate xterms

xterm -hold -e "cd $(dirname $0)/node1; \
go run node1.go -peer localhost:7272 -name agent1" &

xterm -hold -e "cd $(dirname $0)/node2; \
go run node2.go -port 7272 -name agent2" &

xterm -hold -e "cd $(dirname $0)/node3; \
go run node3.go -peer localhost:7272 -name agent3" &

wait
