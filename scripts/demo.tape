# VHS tape file for generating tokenizer demo
# Run with: make demo

Output docs/demo.svg
Output docs/demo.gif

# Terminal appearance
Set FontSize 13
Set FontFamily "JetBrains Mono"
Set Width 900
Set Height 300
Set Padding 15
Set BorderRadius 6
Set WindowBar "Colorful"
Set WindowBarTitle "Terminal"
Set Theme "GitHub Dark"
Set TypingSpeed 50ms

# Practical encoding
Type "# Tokenize Python code"
Enter
Type "echo 'def add(a, b): return a + b' | tokenizer llama3"
Enter
Sleep 2s

# Count tokens
Type "# Count tokens in text"
Enter
Type "echo 'Hello world' | tokenizer llama3 encode --output json | jq length"
Enter
Sleep 2s

# Without special tokens
Type "# Raw token count (no special tokens)"
Enter
Type "tokenizer llama3 encode --bos=false --eos=false 'Hello world'"
Enter
Sleep 2s

Ctrl+L

# Round trip test
Type "# Test encoding/decoding"
Enter
Type "tokenizer llama3 'Test data' | tokenizer llama3 decode"
Enter
Sleep 500ms
Enter
Sleep 2s

# Tokenizer info
Type "# Tokenizer info"
Enter
Type "tokenizer llama3 info | head -8"
Enter
Sleep 2s

Type "# Install: brew install agentstation/tap/tokenizer"
Enter
Sleep 1s