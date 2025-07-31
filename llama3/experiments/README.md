# Experimental Optimizations

This directory contains experimental optimizations that were explored but not included in the main codebase due to compatibility or complexity issues.

## Files

- `state_machine_optimized.go` - Jump table optimization for first-character dispatch
- `state_machine_ascii_opt.go` - ASCII fast paths for character classification  
- `state_machine_pattern_opt.go` - Pattern-specific optimizations
- Various benchmark files for testing these optimizations

## Results

These optimizations showed performance improvements in some cases but had compatibility issues:

- Jump table: ~18% faster on ASCII text but difficult to maintain exact regex behavior
- ASCII fast paths: No improvement - Go's unicode package already optimized
- Pattern-specific: Performance gains but compatibility issues with pattern matching order

## Lessons Learned

1. Maintaining exact compatibility with complex regex behavior is challenging
2. Go's standard library is often already well-optimized
3. Memory optimizations provide the best bang for buck
4. Always verify compatibility with comprehensive test suites before optimizing

These experiments remain here for reference and future exploration.