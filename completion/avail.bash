_comp_xfunc_avail_get_proc() {
	local OPTIND=1 OPTARG="" OPTERR=0 _opt
	local -n _out=$1

	while getopts ':p:P:c' _opt "${@:2}"; do
		case $_opt in
		P | p | c)
			_out=("-$_opt")
			if [[ -n $OPTARG ]]; then
				_out+=("$OPTARG")
			fi

			break
			;;
		*) continue ;;
		esac
	done
}

_comp_cmd_avail() {
	# shellcheck disable=SC2034
	local cur prev words cword comp_args
	_comp_initialize -n : -- "$@" || return

	local cmds
	cmds="run status list schema"

	local global_opts run_opts pid_opts status_opts list_opts schema_opts
	global_opts="-h -v"
	run_opts="-h -c"
	pid_opts="-P -p -c"
	status_opts="-h $pid_opts"
	list_opts="-h $pid_opts"
	schema_opts="-h"

	if [[ $cword -eq 1 ]]; then
		_comp_compgen -- -W "$cmds $global_opts"
		return
	fi

	local subcmd="${words[1]}"

	if [[ $cur == -* ]]; then
		case "$subcmd" in
		run) _comp_compgen -- -W "$run_opts" ;;
		status) _comp_compgen -- -W "$status_opts" ;;
		list) _comp_compgen -- -W "$list_opts" ;;
		schema) _comp_compgen -- -W "$schema_opts" ;;
		*) _comp_compgen -- -W "$global_opts" ;;
		esac
		return
	fi

	case "$subcmd" in
	status | list)
		local -a proc=()
		_comp_xfunc_avail_get_proc proc "${words[@]}"
		IFS=$'\n' read -rd '' -a titles <<<"$(
			avail list "${proc[@]}"
		)"
		_comp_compgen -- -W "${titles[*]}"
		;;
	esac
} &&
	complete -F _comp_cmd_avail avail
