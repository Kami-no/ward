# Ward

Simple bot designed to supplement the free version of GitLab.

## Merge Approves

The free version of GitLab has no Merge Approves. Our team required them while has problems with buying it. Bot tracks emoji for MR.

MR to protected branch notifications:

* MR considered good if it has enough :thumbsup:;
* if MR merged and had not enough qualified :thumbsup: (list is set in its config) than MR is considered bad;
* if MR has at least one :thumbsdown: than MR is considered bad;
* good MR will be marked by the bot with :heavy_check_mark:;
* bad MR will be marked by the bot with :x:;
* if bad MR has been merged bot will mark it with :poop: and will notify people from its list;
* there could be several qualified teams to approve MR;
* by default if there is only one team than MR will require at least 2 :thumbsup: and if there are multiple teams then at least 1 :thumbsup: per team (number of votes per team is customizable at the project level in config).

## Old branches

Once a week bot checks its repositories for stale not protected branches that had no changes:

* wipe merged branches;
* 1 week or more - notify the author of the last commit;
* **TODO:**: 1 month or more - wipe the branch.
