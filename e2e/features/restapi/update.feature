Feature: Create, update, and delete workspaces via REST API

  @skip
  Scenario: users can create a new workspace

  Scenario: users can update their workspaces' visibility
    Given An user is onboarded
    And   Default workspace is created for them
    When  The user changes workspace visibility to "community"
    Then  The workspace visibility is updated to "community" 

  @skip
  Scenario: users can not update visibility of non-owned workspaces

  @skip
  Scenario: users can delete owned workspaces

  @skip
  Scenario: users cannot delete default workspace
