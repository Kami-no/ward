# Ward

Simple tool to be a substitute for Merge Approves in cases when for some reasons you're unable to buy GitLab Subscription.

As a bonus feature it has notifications for old branches.

It tracks MRs in GitLab, gets mail addresses from AD by account name and send notifications via mail.

There is a variation about MR approves here depending on team number:
* one team: 2 approves are required;
* multiple teams: 1 approve from each team are required.
