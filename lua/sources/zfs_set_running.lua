-- Initialization of tables to store information
succeeded = {}
failed = {}

-- Retrieve the arguments
args = ...
argv = args["argv"]

-- First argument is the value to set
value = argv[1]

for i=2, #argv do
    zfs_name = argv[i]
    local err = zfs.sync.set_prop(zfs_name, "label:running", value)
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