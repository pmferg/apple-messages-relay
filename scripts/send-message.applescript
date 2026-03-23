on run argv
	if (count of argv) < 2 then
		error "Usage: osascript send-message.applescript <destination> <payload>"
	end if
	
	set targetHandle to item 1 of argv
	set targetText to item 2 of argv

	tell application "Messages"
		set targetService to 1st service whose service type = iMessage
		set targetBuddy to buddy targetHandle of targetService
		send targetText to targetBuddy
	end tell
end run
