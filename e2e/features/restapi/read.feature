Feature: Read workspaces via REST API

  Scenario: list workspaces with just default workspace
    Given An user is onboarded
    And   Default workspace is created for them
    When  The user requests the list of workspaces
    Then  The user retrieves a list of workspaces containing just the default one

  Scenario: get workspaces with just default workspace
    Given An user is onboarded
    And   Default workspace is created for them
    When  The user requests their default workspace
    Then  The user retrieves their default workspace

  @skip
  Scenario: users can see just their workspaces, the ones shared with them, and the publicly visibile ones

  @skip
  Scenario: users can fetch just their own workspaces

  @skip
  Scenario: users can fetch just the workspaces shared with them

  @skip
  Scenario: users can fetch just the public workspaces
