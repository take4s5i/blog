.PHONY: post

post:
	test "$(SLUG)" != ""
	hugo new posts/$$(date "+%Y")/$$(date "+%m")/$(SLUG).md
