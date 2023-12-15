# Migrate data BY company_id

- Table companies 
- Table Company Config 
- Table Channel Config -> + by channel_id=widget 
- Table roles 
- Table Chat Template 
- Table Tags 
- Table Division 
- Table session_categories 
- Table Users - all user BCADemo
- Table Employee Channel 
- Table rooms (insert ke postgresql & Mongodb) 
- Table sessions 
- Table Participants 
- Table Chat Message (insert ke chat message,Report,Mongo) 
- Table Customer Information 
- Table Pin Rooms 
- Table Unavailable Reason
- Table History Change Unavailable Reason 
- Table History Availability Agent (History Availability User) - sampai 2023-12-13 18:05:00 total data 20638
- Report message
- Report session
- Conversation Mongo
- Room Mongo

# Running binary
nohup /home/sujoko/konnek-migration-engine/workers/chat_message_bulk/chat_message_bulk > "chat_message_bulk_$(date '+%Y-%m-%d').log" &
nohup /path/tempat/file/binary > "nama_log_$(date '+%Y-%m-%d').log" &