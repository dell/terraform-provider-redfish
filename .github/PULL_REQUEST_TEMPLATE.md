<!--
Copyright (c) 2020-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Mozilla Public License Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://mozilla.org/MPL/2.0/


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

<!--- See what makes a good Pull Request at : https://github.com/terraform-providers/terraform-provider-aws/blob/master/docs/CONTRIBUTING.md --->

##### SUMMARY
<!--- Describe the change below, including rationale and design decisions -->

<!--- If your PR fully resolves and should automatically close the linked issue, use Closes. Otherwise, use Relates --->
Relates OR Closes #0000

##### ISSUE TYPE
<!--- Pick one below and delete the rest -->
- Bugfix Pull Request
- Docs Pull Request
- Feature Pull Request
- New Module Pull Request

##### COMPONENT NAME
<!--- Write the short name of the module, plugin, task or feature below -->

##### ADDITIONAL INFORMATION
<!--- Include additional information to help people understand the change here -->
<!--- A step-by-step reproduction of the problem is helpful if there is no related issue -->


<!--- Paste verbatim command output below, e.g. before and after your change -->
```paste below

```

#### Release note for [CHANGELOG](https://github.com/terraform-providers/terraform-provider-aws/blob/master/CHANGELOG.md):
<!--
If change is not user facing, just write "NONE" in the release-note block below.
-->

```release-note

```

#### Output from acceptance testing:

<!--
Replace TestAccXXX with a pattern that matches the tests affected by this PR.

For more information on the `-run` flag, see the `go test` documentation at https://tip.golang.org/cmd/go/#hdr-Testing_flags.
-->
```
$ make testacc TESTARGS='-run=TestAccXXX'

...
```
<!--- Please keep this note for the community --->
#### Community Note

* Please vote on this pull request by adding a 👍 [reaction](https://blog.github.com/2016-03-10-add-reactions-to-pull-requests-issues-and-comments/) to the original pull request comment to help the community and maintainers prioritize this request
* Please do not leave "+1" or other comments that do not add relevant new information or questions, they generate extra noise for pull request followers and do not help prioritize the request

<!--- Thank you for keeping this note for the community --->
