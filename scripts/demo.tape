# VHS tape file for generating tokenizer demo
# Run with: vhs scripts/demo.tape

# Setup
Output docs/demo.gif
Output docs/demo.mp4
Output docs/demo.webm

Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Dracula"

# Title screen
Type "# Tokenizer - High-performance tokenizer implementations in Go" Sleep 500ms Enter
Sleep 2s
Clear

# Install tokenizer
Type "# Install via Homebrew" Sleep 500ms Enter
Sleep 500ms
Type "brew install agentstation/tap/tokenizer" Sleep 500ms Enter
Sleep 5s

Type "# Check version" Sleep 500ms Enter
Type "tokenizer version" Sleep 500ms Enter
Sleep 2s

Clear

# Basic encoding/decoding
Type "# Encode text to tokens" Sleep 500ms Enter
Sleep 500ms
Type "tokenizer llama3 'Hello, world!'" Sleep 500ms Enter
Sleep 2s

Type "# Decode tokens back to text" Sleep 500ms Enter
Type "tokenizer llama3 decode 128000 9906 11 1917 59 0 128001" Sleep 500ms Enter
Sleep 2s

Type "# Decode with special tokens shown" Sleep 500ms Enter
Type "tokenizer llama3 decode --show-special 128000 9906 11 1917 59 0 128001" Sleep 500ms Enter
Sleep 2s

Clear

# Advanced features
Type "# Encode with verbose output" Sleep 500ms Enter
Type "tokenizer llama3 encode --verbose 'The quick brown fox jumps over the lazy dog.'" Sleep 500ms Enter
Sleep 3s

Clear

# Tokenizer info
Type "# Get tokenizer information" Sleep 500ms Enter
Type "tokenizer llama3 info | head -15" Sleep 500ms Enter
Sleep 3s

Clear

# Pipe and streaming
Type "# Stream text from stdin" Sleep 500ms Enter
Type "echo 'Tokenizers are amazing!' | tokenizer llama3" Sleep 500ms Enter
Sleep 2s

Type "# Count tokens using JSON output" Sleep 500ms Enter
Type "echo 'This is a test of the tokenizer.' | tokenizer llama3 encode --format json | jq '.token_count'" Sleep 500ms Enter
Sleep 2s

Clear

# Help commands
Type "# Get help and see all options" Sleep 500ms Enter
Type "tokenizer --help" Sleep 500ms Enter
Sleep 3s

Clear

Type "# See llama3 specific options" Sleep 500ms Enter
Type "tokenizer llama3 --help" Sleep 500ms Enter
Sleep 3s

Clear

# Final message
Type "# Thanks for watching! 🚀" Sleep 500ms Enter
Type "# Learn more at: https://github.com/agentstation/tokenizer" Sleep 500ms Enter
Sleep 3s