# wg-ovpn

Magic util that "bridges" Wireguard with OpenVPN without a TUN/TAP interface

Warning: really ugly and unstable code!

# Building

Obtain latest source of OpenVPN ([link](https://openvpn.net/community-downloads/)),
apply patch `tunsetiff.patch` and place the resulting `openvpn` binary in this project's folder.

Then, run `go build` (requires Go 1.17 or later).

# Usage

`./wg-ovpn <.ovpn file> <wireguard config file>`

Please note that this util doesn't support wg-quick's `.conf` format,
rather it uses `wireguard-go`'s internal UAPI config format:
basically, you can't put section labels like `[Interface]`,
everything else works _roughly_ the same (didn't test though)

Example config:

```
listen_port=55555
private_key=a8dac1d8a70a751f0f699fb14ba1cff7b79cf4fbd8f09f44c6e6a90d0369604f
public_key=28d2b91462b95913ac4fe68259fbabfe4a150314edf04bf4437eaf553d02804c
allowed_ip=0.0.0.0/0
```

# TL;DR How does it work?

It creates a pair of [pseudoterminals](https://man7.org/linux/man-pages/man7/pty.7.html)
that serve as a bidirectional pipe, with one end connected to OpenVPN, and the other one to Wireguard.

It also has a bit of code to replace source/destination IP address to match what OpenVPN expects,
so Wireguard clients can use virtually any IP address and still connect to the OpenVPN network.

# Limitations

- only 1 Wireguard client is currently supported
- I have no idea if it works with other OpenVPN setups than what I have
