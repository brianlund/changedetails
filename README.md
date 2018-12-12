Small go project that can be run as a build step in TeamCity to post changes for the current build to a slack channel.

You'll need to set the following parameters in the TC build:

env.SLACKWEBHOOKURL: Generate this on https://api.slack.com/incoming-webhooks

env.TCBUILDID: %teamcity.build.id%

env.TCPASS: %system.teamcity.auth.password%

env.TCUSER: %system.teamcity.auth.userId%

env.TCHOST: %teamcity.serverUrl%
