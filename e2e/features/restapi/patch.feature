Feature: Patch workspaces via REST API

  Scenario: users can update their workspaces' visibility
    Given An user is onboarded
    And   Default workspace is created for them
    When  The user patches workspace visibility to "community"
    Then  The workspace visibility is updated to "community" 

