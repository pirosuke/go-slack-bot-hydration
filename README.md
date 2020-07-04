# go-slack-bot-hydration
Hydration management app for Slack.

## Install as service (CentOS)

```
sudo cp configs/slack-bot-hydration.service /etc/systemd/system/
sudo systemctl enable slack-bot-hydration
sudo systemctl start slack-bot-hydration
```
