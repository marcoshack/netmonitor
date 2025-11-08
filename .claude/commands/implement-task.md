---
description: Implement a NetMonitor task following project guidelines
arguments:
  - name: task_file
    description: Path to task file (e.g., docs/tasks/T017-xxx.md)
    required: true
---

# Task Implementation: {{task_file}}

## ⚠️ CRITICAL: Documentation Requirements

You are implementing a task for NetMonitor. **You MUST follow these documentation rules:**

### Where to Document (CRITICAL!)
1. ✅ **CORRECT**: Add Implementation Summary to **{{task_file}}** (the original task file)
2. ❌ **WRONG**: Creating separate files like `T0XX-SUMMARY.md`, `T0XX-implementation.md`, etc.

### Required Steps After Implementation
1. Update acceptance criteria in **{{task_file}}**: Change `[ ]` to `[x]`
2. Add `---` separator at the end of **{{task_file}}**
3. Add `## Implementation Summary` section with all required subsections
4. **DO NOT create any separate summary documents**

---

## Task File Structure

Each task file in `docs/tasks/` follows this structure:

1. **Header**: Task number and title
2. **Overview**: Brief description of what needs to be implemented
3. **Context**: Background information and why this task is needed
4. **Task Description**: Detailed description of the work
5. **Acceptance Criteria**: Checklist of requirements (use `[ ]` for pending, `[x]` for completed)
6. **Technical Specifications**: Data structures, interfaces, file formats, etc.
7. **Implementation Requirements**: Specific technical requirements
8. **Verification Steps**: How to verify the implementation
9. **Dependencies**: Other tasks that must be completed first
10. **Notes**: Additional considerations or future enhancements
11. **Implementation Summary**: Added after task completion (see below)

---

## Implementation Process

### 1. Before Starting
- Read **{{task_file}}** completely
- Check dependencies are satisfied
- Understand acceptance criteria
- Plan the implementation approach
- Create todo list with TodoWrite

### 2. During Implementation
- Follow the acceptance criteria
- Write tests as you implement features
- Keep code quality high
- Document key design decisions

### 3. After Completion

⚠️ **IMPORTANT: Follow these steps in {{task_file}} (the ORIGINAL task file)**

#### Step 1: Update Acceptance Criteria
Mark all acceptance criteria as completed by changing `[ ]` to `[x]`:
```markdown
## Acceptance Criteria
- [x] Feature A implemented
- [x] Feature B implemented
- [x] Tests written
```

#### Step 2: Add Implementation Summary Section
**Location: {{task_file}} (THE ORIGINAL TASK FILE)**

Add a comprehensive "Implementation Summary" section at the **END** of **{{task_file}}** (after a horizontal rule `---`).

**⚠️ DO NOT create a separate summary document!**

This section should include:

```markdown
---

## Implementation Summary

[Brief overview of what was implemented]

### Core Features Implemented

#### 1. Feature Name
- **Location**: [file.go:line-range](../../path/to/file.go#Lstart-Lend)
- Description of feature
- Key implementation details
- Important design decisions

[Repeat for each major feature]

### Thread Safety / Concurrency
[If applicable, describe concurrency model and synchronization mechanisms]

### Interface/API
[If applicable, show the main interface or API that was created]

### Test Coverage

Comprehensive test suite added to [test_file.go](../../path/to/test_file.go):

#### Test Cases
1. ✅ **TestName1** - Description
2. ✅ **TestName2** - Description
[List all tests]

#### Test Results
```
[Paste test output showing all tests passing]
```

### File Structure
```
[Show directory structure of implemented files]
```

### Key Design Decisions

#### 1. Decision Name
[Explain the decision and rationale]

#### 2. Decision Name
[Explain the decision and rationale]

### Usage Examples

#### Example 1: Basic Usage
```go
[Code example]
```

#### Example 2: Advanced Usage
```go
[Code example]
```

### Performance Characteristics
- [List performance characteristics, complexity, etc.]

### Future Enhancements
[List potential improvements for future iterations]

### Integration
[Describe how this integrates with other parts of the system]

### Additional Documentation
[Links to other documentation files if created]
```

---

## Best Practices

### Code Quality
- Write clean, readable code
- Follow Go best practices and idioms
- Add comments for complex logic
- Use meaningful variable and function names

### Testing
- Write tests for all public APIs
- Include edge cases and error conditions
- Test concurrent access when applicable
- Aim for >70% code coverage
- Run tests before marking task complete

### Documentation
- Document all exported functions and types
- Include usage examples in documentation
- Link to relevant files and line numbers
- Keep documentation up to date

### Version Control
- Make atomic commits
- Write clear commit messages
- Reference task number in commits (e.g., "T016: Implement atomic file operations")

### Integration
- Ensure changes don't break existing functionality
- Run full project build before completion
- Verify integration with dependent components
- Update related documentation

---

## Task Status Indicators

Use these indicators in task files:
- `[ ]` - Pending/Not started
- `[x]` - Completed
- `[~]` - In progress (optional)
- `[-]` - Blocked or skipped (with explanation)

---

## ⚠️ CRITICAL REMINDERS

### Implementation Summary Location
- ✅ **CORRECT**: Add to **{{task_file}}** after `---` separator
- ❌ **WRONG**: Creating separate summary files like `Txxx-SUMMARY.md` or `Txxx-implementation.md`

### Required Actions on Task Completion
1. Update acceptance criteria in **{{task_file}}**: `[ ]` → `[x]`
2. Add `---` separator after the Notes section in **{{task_file}}**
3. Add `## Implementation Summary` section with all required subsections to **{{task_file}}**
4. Verify all sections follow the template above
5. **DO NOT create any separate summary documents**

### Documentation Quality
- The Implementation Summary should be detailed enough that someone can understand what was done without reading all the code
- Include code examples and test results to demonstrate functionality
- Link to specific file locations with line numbers for easy navigation
- If you create additional documentation (like API references), link to it in the Implementation Summary

---

## Example Reference

See [T016: JSON Storage System](../../docs/tasks/T016-json-storage-system.md) for a complete example of a properly documented completed task.

---

**Now implement the task at {{task_file}}. Start by reading the file and creating a todo list with TodoWrite.**
