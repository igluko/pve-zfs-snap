package main

// return lua code for snapshot
func lua_unset_running() string {
	return `
-- Initialization of tables to store information
succeeded = {}
failed = {}

-- Retrieve the arguments
args = ...
argv = args["argv"]

for i=1, #argv do
    zfs_name = argv[i]
    local err = zfs.sync.inherit(zfs_name, "label:running")
    if (err ~= 0) then
        failed[zfs_name] = err
    else
        succeeded[zfs_name] = err
    end
end

-- Return the results
results = {}
results["succeeded"] = succeeded
results["failed"] = failed
return results
`
}
