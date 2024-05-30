Feature: Create workspaces via REST API

  @skip
  Scenario: users can create a new workspace

  Scenario: user requests a private workspace
    Given An user is onboarded
    When  The user requests a new private workspace
    Then  A private workspace is created
  
  Scenario: user requests a community workspace
    Given An user is onboarded
    When  The user requests a new community workspace
    Then  A community workspace is created

