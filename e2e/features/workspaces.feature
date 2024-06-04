Feature: Workspace lifecycle

  Scenario: user onboarding
    When An user onboards
    Then Default workspace is created for them
    And  The workspace visibility is set to "private"

  Scenario: role is set on workspace for owner user 
    When A workspace is created for an user
    Then The owner is granted admin access to the workspace
