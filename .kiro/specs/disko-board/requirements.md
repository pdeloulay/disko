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
- Admin decides what Idea fields are visible in the board (onliner is always required)

### Definitions

An Idea contains a one-liner, a description, a value statement and a RICE score composed of R, I, C, E. R and I are 0 to 100%, R is considered 1 (hours), 2 (days), 4 (weeked), 8 (months).

An idea can be associated to a specific columns Parking, Now, Next, Later, Release, Wont do

ongoing ideas lands usually on Now, Next, Later - A simple 3 column based layout where ideas can float from one to the other. Idea can be marked as in progress by solopreneur. Idea can be marked as Done, in which case they go the the Release Tab, a table that lists all released ideas.

A new idea will start in Parking Area. Parking Area is a table that lists all ideas that are not in progress. 

Ideas can be moved back to Parking Area

Ideas in progress have a slight animation to them.

Admin can drag and drop ideas from the different columns at any time.

A board is always visible from the internet as long as you get the public link. Public link are automatically generated. 

In Release Mode (via Release Tab)
There is a search and filter to find delivered ideas (name, description)

Each idea can receive feedback thumb up from customers. 

Each board contains a master search where dynamic keyboard search only display in board search results. Each column can be sorted by State (In Progress or not, RICE score, by Name)

### Access

Access via App landing page on product hero, benefits, testimonials, description and sign-in, sign-up buttons for Board Owners. Indicate how many boards are created via a incremental ticker visualization. 
Access to board via dashboard for authentication Board owners.
Access to public boards from the Internet via shared link.

### UI

- Dark mode is available on all pages
- Visitors cam swap Dark/White mode on marketing and app pages

## Requirements

### Requirement 1: Board Management

**User Story:** As an Admin, I want to create and manage boards, so that I can organize and share my work progress with customers.

#### Acceptance Criteria

1. WHEN an Admin creates a new board THEN the system SHALL generate a unique public link automatically
2. WHEN an Admin deletes a board THEN the system SHALL delete all ideas contained in that board
3. WHEN an Admin accesses their dashboard THEN the system SHALL display all boards they have created
4. WHEN a board is created THEN the system SHALL initialize it with default columns (Parking, Now, Next, Later, Release, Won't do)

### Requirement 2: Idea Management

**User Story:** As an Admin, I want to create, edit, and manage ideas within my boards, so that I can track my work items and their progress.

#### Acceptance Criteria

1. WHEN an Admin creates a new idea THEN the system SHALL place it in the Parking Area by default
2. WHEN an Admin creates an idea THEN the system SHALL require a one-liner, description, value statement, and RICE score (R: 0-100%, I: 0-100%, C: 1/2/4/8, E: 0-100%)
3. WHEN an Admin edits an idea THEN the system SHALL update the idea information and maintain its current column position
4. WHEN an Admin deletes an idea THEN the system SHALL remove it from the board permanently
5. WHEN an Admin marks an idea as "in progress" THEN the system SHALL apply a visual animation to indicate active status
6. WHEN an Admin marks an idea as "Done" THEN the system SHALL move it to the Release column automatically

### Requirement 3: Drag and Drop Functionality

**User Story:** As an Admin, I want to drag and drop ideas between columns, so that I can easily update their status and priority.

#### Acceptance Criteria

1. WHEN an Admin drags an idea from one column to another THEN the system SHALL update the idea's status to match the target column
2. WHEN an Admin drops an idea in a valid column THEN the system SHALL save the new position immediately
3. WHEN an Admin drags an idea THEN the system SHALL provide visual feedback during the drag operation
4. WHEN an Admin moves an idea back to Parking Area THEN the system SHALL remove any "in progress" status

### Requirement 4: Public Board Access

**User Story:** As a Public user, I want to view boards via public links, so that I can see what the entrepreneur is working on.

#### Acceptance Criteria

1. WHEN a Public user accesses a board via public link THEN the system SHALL display the board with visible columns only
2. WHEN a Public user views a board THEN the system SHALL show ideas with their one-liner, description, and value statement
3. WHEN a Public user views ideas THEN the system SHALL NOT display RICE scores or admin-only information
4. WHEN a Public user accesses an invalid or deleted board link THEN the system SHALL display an appropriate error message

### Requirement 5: Public Feedback System

**User Story:** As a Public user, I want to provide feedback on ideas through thumbs up and emojis, so that I can show support for the entrepreneur's work.

#### Acceptance Criteria

1. WHEN a Public user clicks thumbs up on an idea THEN the system SHALL increment the thumbs up counter
2. WHEN a Public user selects an emoji for an idea THEN the system SHALL record the emoji feedback
3. WHEN a Public user provides feedback THEN the system SHALL update the feedback display in real-time
4. WHEN a Public user tries to provide feedback multiple times THEN the system SHALL allow multiple interactions per user
5. WHEN a Public user tries to provide feedback, the system should protect itself from repetitive actions by the same user, same IP source
6. WHEN a Public user provides feedback, notification is sent in real tine to the Admin via multi-channel notifications, Email, Slack, Web Hooks, and Boards. Animation should be visible on the board if visible

### Requirement 6: Column Visibility Control

**User Story:** As an Admin, I want to control which columns are visible on my board, so that I can customize the public view.

#### Acceptance Criteria

1. WHEN an Admin toggles column visibility THEN the system SHALL show/hide the column for both Admin and Public views
2. WHEN an Admin hides a column containing ideas THEN the system SHALL keep the ideas but make them invisible to Public users
3. WHEN an Admin makes a column visible THEN the system SHALL immediately display it with all contained ideas

### Requirement 7: Release Tab and Search

**User Story:** As both Admin and Public users, I want to view and search completed ideas in the Release tab, so that I can see what has been delivered.

#### Acceptance Criteria

1. WHEN a user accesses the Release tab THEN the system SHALL display all ideas marked as "Done" in a table format
2. WHEN a user searches in the Release tab THEN the system SHALL filter results by name and description
3. WHEN a user applies filters in the Release tab THEN the system SHALL update the displayed results accordingly
4. WHEN ideas are moved to Release THEN the system SHALL maintain their feedback (thumbs up, emojis)

### Requirement 8: Board Search and Sorting

**User Story:** As a user, I want to search and sort ideas within the board, so that I can quickly find specific items.

#### Acceptance Criteria

1. WHEN a user types in the master search THEN the system SHALL dynamically filter and display matching ideas across all visible columns
2. WHEN a user sorts a column by "In Progress" status THEN the system SHALL group in-progress ideas at the top
3. WHEN a user sorts a column by RICE score THEN the system SHALL order ideas from highest to lowest score
4. WHEN a user sorts a column by name THEN the system SHALL order ideas alphabetically
5. WHEN search results are displayed THEN the system SHALL highlight matching text in idea titles and descriptions

### Requirement 9: Application Landing Page and Access

**User Story:** As a potential Board Owner, I want to access the application through a marketing landing page, so that I can understand the product and sign up.

#### Acceptance Criteria

1. WHEN a visitor accesses the application root URL THEN the system SHALL display a landing page with product hero, benefits, testimonials, and description
2. WHEN a visitor views the landing page THEN the system SHALL display sign-in and sign-up buttons for Board Owners
3. WHEN a visitor views the landing page THEN the system SHALL show an incremental ticker visualization indicating how many boards have been created
4. WHEN an authenticated Board Owner accesses the application THEN the system SHALL redirect them to their dashboard
5. WHEN a Board Owner signs up or signs in THEN the system SHALL authenticate them via Clerk and redirect to their dashboard

### Requirement 10: Dark Mode Support

**User Story:** As a user, I want to toggle between dark and light modes across all pages, so that I can use the application in my preferred visual theme.

#### Acceptance Criteria

1. WHEN a user accesses any page of the application THEN the system SHALL provide a dark mode toggle option
2. WHEN a user toggles to dark mode THEN the system SHALL apply dark theme styling to all UI components
3. WHEN a user toggles to light mode THEN the system SHALL apply light theme styling to all UI components
4. WHEN a user changes theme preference THEN the system SHALL persist the preference across browser sessions
5. WHEN a user accesses the application THEN the system SHALL load their previously selected theme preference
6. WHEN a public user accesses a board via shared link THEN the system SHALL provide theme toggle functionality
## Notes

- [Any additional context, constraints, or considerations]
- [Technical limitations or dependencies]
- [Performance requirements if applicable]