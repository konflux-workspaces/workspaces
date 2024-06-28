# End-To-End tests
e2e tests are implemented following the [Behavior-Driven Development (BDD) approach][bdd-approach].

Implementation relies on [cucumber/godog][cucumber-godog].

* the [`assets` folder][e2e-assets-folder] contains the assets needed to run the e2e tests, like external CRDs
* the [`feature` folder][e2e-feature-folder] contains the BDD Feature files describing the scenarios to test
* the [`hook` folder][e2e-hook-folder] contains the godog hooks that are executed before and after each suite/test/step.
* the [`step` folder][e2e-step-folder] contains the Go code implementing the steps.
    Steps are organized by domain and then for step type (given, when, then, or step).


<!-- external links -->

[e2e-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/e2e
[e2e-assets-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/e2e/assets
[e2e-feature-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/e2e/feature
[e2e-hook-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/e2e/hook
[e2e-step-folder]: https://github.com/konflux-workspaces/workspaces/tree/main/e2e/step

[bdd-approach]: https://en.wikipedia.org/wiki/Behavior-driven_development
[cucumber-godog]: https://github.com/cucumber/godog
