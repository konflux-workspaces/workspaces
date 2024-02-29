Feature: REST API

  Scenario: list workspaces with just default workspace
    Given An user is onboarded
    And   Default workspace is created for them
    When  The user requests the list of workspaces
    Then  The user retrieves a list of workspaces containing just the default one

  @wip
  Scenario: get workspaces with just default workspace
    Given An user is onboarded
    And   Default workspace is created for them
    When  The user requests their default workspace
    Then  The user retrieves their default workspace

