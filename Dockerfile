# syntax=docker/dockerfile:1
FROM scratch

# Copy the binary
COPY tokenizer /tokenizer

# Set the binary as entrypoint
ENTRYPOINT ["/tokenizer"]