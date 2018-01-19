package ergonode

// Distributed operations codes (http://www.erlang.org/doc/apps/erts/erl_dist_protocol.html)
const (
	LINK         = 1
	SEND         = 2
	EXIT         = 3
	UNLINK       = 4
	NODE_LINK    = 5
	REG_SEND     = 6
	GROUP_LEADER = 7
	EXIT2        = 8
	SEND_TT      = 12
	EXIT_TT      = 13
	REG_SEND_TT  = 16
	EXIT2_TT     = 18
	MONITOR      = 19
	DEMONITOR    = 20
	MONITOR_EXIT = 21
)
