package templates

// SSHConfig stores the ssh config template.
var SSHConfig = `{{range .BastionHosts}}# Bastion - {{.Region}}{{if .ID}} #{{.ID}}{{end}}
Host {{.Region}}{{if .ID}}-0{{.ID}}{{end}}
HostName {{.IP}}
StrictHostKeyChecking no

{{end}}# ProxyJump configuration
Host 172.*
ProxyJump {{.JumpHost}}
StrictHostKeyChecking no`
