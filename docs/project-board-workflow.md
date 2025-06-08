# GitHub Project Board Workflow

## Overview

This document defines the workflow and usage guidelines for the Traffic Control Go development project board, which serves as the central hub for task management and project visibility.

## Board Structure

### Columns and Workflow

#### 1. 📋 Backlog
**Purpose**: New issues awaiting prioritization  
**Entry Criteria**: 
- New issue created
- Not yet prioritized or assigned

**Exit Criteria**:
- Issue has been prioritized (priority label added)
- Ready for development planning
- Move to: Ready

#### 2. 🔍 Ready
**Purpose**: Prioritized issues ready to start  
**Entry Criteria**:
- Clear requirements defined
- Priority label assigned
- No blocking dependencies

**Exit Criteria**:
- Developer assigned
- Work is about to begin
- Move to: In Progress

#### 3. 🚧 In Progress
**Purpose**: Active development work  
**Entry Criteria**:
- Developer actively working on issue
- Branch created for implementation
- WIP limit: 3 issues per developer

**Exit Criteria**:
- Implementation complete
- Pull request created
- Move to: In Review

#### 4. 👀 In Review
**Purpose**: Code review and feedback  
**Entry Criteria**:
- Pull request created and ready for review
- CI checks passing
- Linked to issue

**Exit Criteria**:
- Code review approved
- All feedback addressed
- Move to: Testing or Done

#### 5. 🧪 Testing
**Purpose**: Additional testing/validation  
**Entry Criteria**:
- Complex features requiring extra validation
- Integration testing needed
- Performance validation required

**Exit Criteria**:
- All tests passing
- Performance validated
- Move to: Done

#### 6. ✅ Done
**Purpose**: Completed work  
**Entry Criteria**:
- PR merged to main branch
- Issue closed automatically
- All acceptance criteria met

**Archive Policy**:
- Items remain for 1 sprint/iteration
- Then archived for historical tracking

## Workflow Rules

### Work In Progress (WIP) Limits
- **In Progress**: Maximum 3 issues per developer
- **In Review**: Maximum 5 PRs team-wide
- **Testing**: Maximum 3 items

### Daily Workflow
1. **Morning**: Review board status, update issue positions
2. **During Work**: Move issues as status changes
3. **End of Day**: Ensure board reflects current state

### Weekly Activities
- **Monday**: Board grooming, prioritization
- **Wednesday**: Mid-week progress check
- **Friday**: Week wrap-up, planning ahead

## Automation Rules

### Automatic Transitions
1. **Issue Created** → Automatically added to Backlog
2. **PR Created** → Linked issue moves to In Review
3. **PR Merged** → Issue moves to Done and closes

### Label-Based Automation
- `blocked` label → Visual indicator on board
- `critical-priority` → Top of Ready column
- `bug` → Fast-track through workflow

## Board Queries and Views

### Priority View
```
is:open label:critical-priority
is:open label:high-priority
is:open label:medium-priority
is:open label:low-priority
```

### Milestone View
```
is:open milestone:"Milestone 1: Project Foundation"
is:open milestone:"Milestone 2: Core API Development"
is:open milestone:"Milestone 3: Advanced Features"
```

### Team Member View
```
is:open assignee:@username
is:open reviewer:@username
```

## Visual Indicators

### Card Information
- **Title**: Issue title with number
- **Labels**: Priority, type, and status labels
- **Assignee**: Developer responsible
- **Milestone**: Current development phase
- **Due Date**: If applicable

### Status Indicators
- 🔴 **Blocked**: Red indicator for blocked items
- 🟡 **At Risk**: Yellow for items approaching deadline
- 🟢 **On Track**: Green for normal progress

## Metrics and Reporting

### Key Metrics to Track
1. **Cycle Time**: Time from Ready to Done
2. **Lead Time**: Time from Backlog to Done
3. **WIP Compliance**: Adherence to WIP limits
4. **Throughput**: Items completed per week

### Weekly Metrics Review
- Average cycle time by issue type
- Bottleneck identification
- Team velocity trends
- Blocked item analysis

## Best Practices

### For Developers
1. Update issue status in real-time
2. Add blockers immediately when identified
3. Link PRs to issues consistently
4. Follow WIP limits strictly

### For Project Managers
1. Daily board review for blockers
2. Weekly metrics analysis
3. Regular backlog grooming
4. Milestone progress tracking

### For Stakeholders
1. View project board for real-time status
2. Check Done column for recent completions
3. Review metrics for project health

## Integration with Other Tools

### GitHub Issues
- All work items tracked as issues
- Labels for categorization
- Milestones for phase tracking

### Pull Requests
- Linked to issues via "Fixes #XXX"
- Automatic status updates
- Review status visible

### GitHub Actions
- CI/CD status on PR cards
- Automated testing results
- Deployment status tracking

## Troubleshooting

### Common Issues
1. **Card Stuck in Column**: Check for missing PR link or unresolved blockers
2. **Automation Not Working**: Verify PR-issue linking syntax
3. **WIP Limit Exceeded**: Review and complete in-progress items

### Board Maintenance
- Weekly cleanup of Done column
- Monthly review of automation rules
- Quarterly workflow optimization

This workflow ensures efficient task management, clear visibility, and consistent project progress tracking throughout the development lifecycle.