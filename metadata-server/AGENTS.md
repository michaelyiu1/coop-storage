# AGENTS.md

This serves as the metadata storage for the adjacent object storage. This server is used to retrieve user and object metadata, allowing the client to do CRUD on the object storage.

# badgerdb schema

The BadgerDB kv store is structured into different indexes. Here are the general schemas for the keys and values (the curly brackets indicate the infix, like a formatted string):

## Object

Keys: "objid:{objId}"
- Note:  objId is the object id in the object storage device.

Value: { ownerId : num, 
      sharedWith : bool,
      FileType: string,
      FileName: string,
       DeleteFlag: bool,
      Version: string,
      # TODO: previewId, Size, LastModified }

## User home directory
Key: "userid:{userId}:~"

Value: []string | FileNametoObejctID
  e.g., ["dir1/", "dir2/", {"file_name1": "obj12345"}]

## User sub directories

Key: "userid:{userId}:~/{subDir}"
- Note: these keys can be arbitrarily nest e.g., "userid:12345:~/dir1/dir2A"

Value: []string | FileNametoObejctID
  e.g., ["dir1A/", {"file_name2":"871263"]

  

