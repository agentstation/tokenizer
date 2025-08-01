# VHS tape file for generating tokenizer demo GIF
# Run with: vhs scripts/demo.tape

# Setup
Output docs/demo.gif
Set FontSize 14
Set Width 800
Set Height 600
Set Theme "Monokai Pro"

# Start recording
Type "# Install Tokenizer CLI" Sleep 500ms Enter
Sleep 1s
Type "brew install agentstation/tap/tokenizer" Sleep 500ms Enter
Sleep 2s

Type "tokenizer version" Sleep 500ms Enter
Sleep 2s

Ctrl+L

Type "# Encode text with Llama 3" Sleep 500ms Enter
Sleep 1s
Type "tokenizer llama3 'Hello, world!'" Sleep 500ms Enter
Sleep 2s

Type "# Decode tokens back to text" Sleep 500ms Enter
Type "tokenizer llama3 decode 128000 9906 11 1917 0 128001" Sleep 500ms Enter
Sleep 2s

Ctrl+L

Type "# Get tokenizer info" Sleep 500ms Enter
Type "tokenizer llama3 info" Sleep 500ms Enter
Sleep 3s

Ctrl+L

Type "# Stream processing" Sleep 500ms Enter
Type "echo 'The quick brown fox' | tokenizer llama3 stream" Sleep 500ms Enter
Sleep 2s

Type "# Direct encoding" Sleep 500ms Enter
Type "tokenizer llama3 'I love tokenizers!'" Sleep 500ms Enter
Sleep 2s