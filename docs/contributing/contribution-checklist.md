# Contribution Checklist

- Your PR has been squashed to a single commit
- Any new functionality has accompanying unit test cases
- Make sure any pull requests to the repo rebase cleanly without conflict on master
- Commits should be signed off with your name in the format `Signed-off-by: FirstName LastName <gelante@someemail.domain>` at the end. A good commit looks like this:

        file_I_modified: General description of change <---- this is the title
        - If you didn't modify one specific file then use a descriptive message for the title
        - Another description of a change you made
        - A third description of a change you made

        Signed-off-by: FirstName LastName <gelante@someemail.domain>
- If your code is addressing an issue use the `Fixes #<ticket_number>` syntax inside the commit message to reference
  the ticket. This will automatically close the ticket when the PR is accepted
- Branches preferably follow GitHub's naming syntax for addressing tickets. For example, a ticket called *Unit test
  cases for power control #25* would have a branch called *25-unit-test-cases-for-power-control*.
- All functions aside from the context functions must have comments. At a minimum a short sentence explaining what it does
- 