package portable

import (
	"net"
	"net/url"
	"regexp"
	"strings"
)

var sshName = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.+-]*$`)
var sshHost = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]*$`)

// SSH authentication and host trust remain machine-local OpenSSH concerns.
// Accept Git's SSH URL and common scp forms, never remote-helper syntax,
// embedded passwords, option-like hosts/users or control characters.
func isSSHRemote(remote string) bool {
	if strings.ContainsAny(remote, "\x00\r\n\t") {
		return false
	}
	if strings.HasPrefix(remote, "ssh://") {
		u, err := url.Parse(remote)
		if err != nil || u.Host == "" || u.Path == "" || u.Path == "/" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
			return false
		}
		if !sshHost.MatchString(u.Hostname()) && net.ParseIP(u.Hostname()) == nil {
			return false
		}
		if u.User != nil {
			if _, password := u.User.Password(); password || !sshName.MatchString(u.User.Username()) {
				return false
			}
		}
		return !strings.ContainsAny(u.Path, "\x00\r\n\t")
	}
	if strings.Contains(remote, "://") {
		return false
	}
	endpoint, path, found := strings.Cut(remote, ":")
	if !found || path == "" || strings.HasPrefix(path, ":") || strings.HasPrefix(path, "-") {
		return false
	}
	if user, host, ok := strings.Cut(endpoint, "@"); ok {
		return sshName.MatchString(user) && sshHost.MatchString(host)
	}
	return sshHost.MatchString(endpoint)
}
