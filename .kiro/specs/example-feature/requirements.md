# Requirements Document

## Introduction

Disko in a simple app for entrepreneur to share with their customers what they are currently working on. We call it a Disko Board.

### Roles

- Admin Role - Solopreneurs as playing Admin role can manage boards. Boards contains ideas that they are working on. Ideas are also managed by Admin/Owner.
- Public Role - FOlks with the board public link can view the board and ideas. They can only thumb up or emoji any visible idea on the board.

#### Admin Role

- Admin can create board(s)
- Admin can delete board(s)
- Admin can delete a board but will delete all ideas in that board
- Admin can add, edit, delete ideas in a board
- Admin decides with column is visible in the Board

### Definitions

An Idea contains a one-liner, a description, a value statement and a RICE score composed of R, I, C, E. R and I are 0 to 100%, R is considered 1 (hours), 2 (days), 4 (weeked), 8 (months).

An idea can be associated to a specific columns Parking, Now, Next, Later, Release, Wont do

Ongoing ideas lands usually on Now, Next, Later - A simple 3 column based layout where ideas can float from one to the other. Idea can be marked as in progress by solopreneur. Idea can be marked as Done, in which case they go the the Release Tab, a table that lists all released ideas.

A new idea will start in Parking Area. Parking Area is a table that lists all ideas that are not in progress. 

Ideas can be moved back to Parking Area

Ideas in progress have a slight animation to them.

Solopreneur or Admin can drag and drop ideas from the different columns at any time.

A board is always visible from the internet as long as you get the public link. Public link are automatically generated. 

In Release Mode (via Release Tab)
There is a search and filter to find delivered ideas (name, description)

Each idea can receive feedback thumb up from customers. 

Each board contains a master search where dynamic keyboard search only display in board search results. Each column can be sorted by State (In Progress or not, RICE score, by Name)

## User Interface

Board are managed on a single page with minimal overlay for edits and Delete.
Boards have column and contains ideas
Idea can be drag and dropped.


## Requirements

### Requirement 1: [Feature Name]

**User Story:** As a [role], I want [feature], so that [benefit]

#### Acceptance Criteria

1. WHEN [event/trigger] THEN [system] SHALL [response/behavior]
2. IF [precondition] THEN [system] SHALL [response/behavior]
3. WHEN [event] AND [condition] THEN [system] SHALL [response/behavior]

### Requirement 2: [Another Feature Aspect]

**User Story:** As a [role], I want [feature], so that [benefit]

#### Acceptance Criteria

1. WHEN [event/trigger] THEN [system] SHALL [response/behavior]
2. IF [precondition] THEN [system] SHALL [response/behavior]
3. WHEN [event] AND [condition] THEN [system] SHALL [response/behavior]

### Requirement 3: [Error Handling/Edge Cases]

**User Story:** As a [role], I want [error handling behavior], so that [benefit]

#### Acceptance Criteria

1. WHEN [error condition] THEN [system] SHALL [error response]
2. IF [invalid input] THEN [system] SHALL [validation response]
3. WHEN [system failure] THEN [system] SHALL [fallback behavior]

## Notes

- [Any additional context, constraints, or considerations]
- [Technical limitations or dependencies]
- [Performance requirements if applicable]