package templates

// SSHConfig stores the ssh config template.
var SSHConfig = `{{range .BastionHosts}}# Bastion - {{.Region}} #{{.ID}}
Host {{.Region}}-0{{.ID}}
HostName {{.IP}}
StrictHostKeyChecking no

{{end}}# ProxyJump configuration
Host 172.*
ProxyJump {{.JumpHost}}
StrictHostKeyChecking no`
