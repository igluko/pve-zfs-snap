package main

// return lua code for snapshot
func lua_destroy() string {
	return `-- Initialization of tables to store information about the created snapshots and any errors that occurred
    succeeded = {}
    failed = {}
    
    -- Retrieve the arguments
    args = ...
    argv = args["argv"]
    
    -- Create a snapshot for each provided filesystem
    for i=1, #argv do
        snap_name = argv[i]
        local err = zfs.sync.destroy(snap_name)
        if (err ~= 0) then
            failed[snap_name] = err
        else
            succeeded[snap_name] = err
        end
    end
    
    -- Return the results
    results = {}
    results["succeeded"] = succeeded
    results["failed"] = failed
    return results    
`
}
