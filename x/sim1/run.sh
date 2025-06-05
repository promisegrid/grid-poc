#!/usr/bin/env bash

# Check if tmux is installed
if ! command -v tmux >/dev/null 2>&1; then
    echo "Error: tmux is required but not installed." >&2
    exit 1
fi

# Create new tmux session with initial window for node1
tmux new-session -d -s sim1 -c node1 -n nodes \
    "go run node1.go -peer localhost:7272 -name agent1"
tmux select-pane -T "node1"

# Split window for node2 with working directory
tmux split-window -c node2 -t sim1 \
    "go run node2.go -port 7272 -name agent2"
tmux select-pane -T "node2"

# Split window for node3 with working directory  
tmux split-window -c node3 -t sim1 \
    "go run node3.go -peer localhost:7272 -name agent3"
tmux select-pane -T "node3"

# Arrange panes evenly and attach session
tmux select-layout -t sim1 even-vertical
tmux attach-session -t sim1
