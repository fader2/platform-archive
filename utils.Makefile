
findpid:
	ps -ax | grep -i platform
.PHONY: findpid

reload:
	@echo "kill -SIGUSR1 PID"
.PHONY: reload