# Compilation Fixes Summary

This document summarizes the fixes applied to resolve CI compilation errors in the traffic-control-go project.

## 1. Event Store Type Error

**Issue**: `events.DomainEvent is not a type` in `internal/infrastructure/eventstore/memory.go`

**Cause**: The parameter name `events` in the `Save` method was shadowing the imported package name `events`.

**Fix**: Renamed the parameter from `events` to `domainEvents` to avoid the naming conflict.

## 2. Netlink Compatibility Issues

### 2.1 FwFilter Mark Field

**Issue**: `unknown field Mark in struct literal of type FwFilter`

**Cause**: The vishvananda/netlink v1.3.1 FwFilter struct doesn't have a `Mark` field.

**Fix**: Removed the Mark field assignment and added a comment indicating that mark-based filtering would require using U32 filter or updating the netlink library.

### 2.2 Netem Fields and Type Mismatches

**Issues**:
- `netem.Corrupt undefined` and `netem.Reorder undefined`
- Type mismatches between `float32` and `uint32` for Loss, Duplicate, etc.

**Cause**: 
- The fields are named `CorruptProb` and `ReorderProb`, not `Corrupt` and `Reorder`
- These fields expect uint32 values representing percentages in kernel's fixed-point format

**Fix**: 
- Changed field names to `CorruptProb` and `ReorderProb`
- Added conversion from float32 percentage (0-100) to kernel's uint32 representation

### 2.3 Statistics Interface Issues

**Issue**: `invalid operation: qdisc.Attrs().Statistics is not an interface`

**Cause**: The Statistics field is a concrete type `*QdiscStatistics` (which is an alias for `ClassStatistics`), not an interface.

**Fix**: Removed the type assertion and accessed the nested fields directly through the ClassStatistics structure.

### 2.4 Missing HTB Class Stats

**Issue**: `htbClass.Stats undefined`

**Cause**: The HtbClass struct in netlink v1.3.1 doesn't have a Stats field.

**Fix**: Added a comment explaining the limitation and set HTB-specific statistics to zero values.

## 3. Examples Package Issue

**Issue**: `main redeclared in this block`

**Cause**: Multiple files in the examples directory had `package main` with `func main()`.

**Fix**: Added `// +build ignore` build tag to both example files to exclude them from regular builds.

## 4. Type Conversion Issue

**Issue**: `conversion from QdiscType (int) to string yields a string of one rune`

**Cause**: Attempting to convert QdiscType directly to string using `string(config.Type)`.

**Fix**: Used the String() method: `config.Type.String()`.

## 5. Unused Import

**Issue**: `"github.com/rng999/traffic-control-go/internal/domain/entities" imported and not used`

**Fix**: Removed the unused import from netem.go.

## Notes on Netlink Library Limitations

The current version of vishvananda/netlink (v1.3.1) has several limitations:

1. **FwFilter**: Doesn't support Mark field for firewall mark-based filtering
2. **Statistics**: Limited statistics available compared to what the kernel provides
3. **HTB Class Stats**: Detailed HTB statistics (lends, borrows, etc.) are not exposed

To get full functionality, consider:
- Updating to a newer version of the netlink library if available
- Using the tc command directly for advanced features
- Implementing custom netlink message handling for unsupported features