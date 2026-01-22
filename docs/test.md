aimgr repo add skill <skill-expression> //same as the current ai-repo add
aimgr repo remove skill <skill-name> //same as the current ai-repo remove
aimgr repo update skill <skill-name> //update the content in the repo, resync from the file or from the github repo
aimgr repo update skill <updates all skills>
aimgr repo update //updates all skills, agents, commands

aimgr repo list //same as ai-rep list right now
aimgr repo list skill //list only the skills -> should work the same for command and agent
aimgr repo show skill <skill-name> //shows details about the skill, name, description github url or filesystem url, last update timestamp, ...


aimgr install skill/<skill-name> //same as now ai-repo install skill dynatrace-control
aimgr install skill/<skill-name1>  skill/<skill-name1> command/<command-name> //now we support multiple installs in once command  

aimgr uninstall skill/<skill-name> //uninstall the installed skill only do anything if this is a symlink to our repo
aimgr uninstall skill/<skill-name1> skill/<skill-name2> //se support  multiple uninstalls in one command 


all this stuff must work for all types agents skills and commands
