Package swdeploy

The swdeploy service accepts a command via etcd indicating which version of the 'shell' repo should be checked out. It then checks out that version o f the shell and does a mr update using the .mrconfig_production file of the shell. This will update all repos to their specified version defined in said file. Then the service runs a script named 'deploy' located at the top of each repo tree. All the legwork is thus done in the 'deploy' script. Systemd services should also be installed in this script and it goes without saying, the script must be idempotent.



