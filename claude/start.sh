#!/run/current-system/sw/bin/bash
docker run \
	--cap-add=NET_ADMIN \
	--cap-add=NET_RAW \
	--env NODE_OPTIONS=--max-old-space-size=4096 \
	--env CLAUDE_CONFIG_DIR=/home/node/.claude \
	--env POWERLEVEL9K_DISABLE_GITSTATUS=true \
	-it \
	--mount "source=claude-code-bashhistory,target=/commandhistory,type=volume" \
	--mount "source=cloude-code-config,target=/home/node/.claude,type=volume" \
	--mount "source=./,target=/workspace,type=bind,consistency=delegated" \
	--name claude \
	--rm claudecode:latest claude --dangerously-skip-permissions
