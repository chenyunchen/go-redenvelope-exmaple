--
-- Created by IntelliJ IDEA.
-- User: chenyunchen
-- Date: 2019-03-13
-- Time: 11:51
-- To change this template use File | Settings | File Templates.
--

local len = 100

local function getRedenvelope(rID, userID)
    if redis.call('hexists', rID, userID) ~= 0 then
        return nil;
    else
        local redenvelope = redis.call('rpop', 'redenvelope');
        if redenvelope then
            local x = cjson.decode(redenvelope);
            x['userID'] = tonumber(userID);
            local re = cjson.encode(x);
            redis.call('hset', rID, userID, userID);
            redis.call('lpush', 'consume_' .. rID, re);
            return re;
        end
    end
    return nil;
end

local function getHeader(name)
    return redis.call('getrange', name, 0, 7);
end

local function parseHeader(header)
    return struct.unpack('I4HH', header);
end

local function setHeader(name, lastID, head, tail)
    local header = struct.pack('I4HH', lastID, head, tail);
    redis.call('setrange', name, 0, header);
end

local function push(name, head, tail, id)
    local idStr = struct.pack('I4', id);
    redis.call('setrange', name, tail + 8, idStr);
    tail = (tail + 4) % len;
    if tail == head then
        head = (head + 4) % len;
    end
    return head, tail;
end

local function pushMsg(name, msg, msgID)
    local lastID, head, tail = 0, 0, 0;
    local header = getHeader(name);
    if header ~= '' then
        lastID, head, tail = parseHeader(header);
    end

    local mID = tonumber(msgID);
    if mID <= lastID then mID = lastID + 1 end

    redis.call('set', 'msg:' .. mID, msg);
    local head, tail = push(name, head, tail, mID);
    setHeader(name, mID, head, tail);
    return {msgID, lastID};
end

local function getMsgs(name)
    local lastID, head, tail = 0, 0, 0;
    local header = getHeader(name);
    if header == '' then
        return {};
    end
    lastID, head, tail = parseHeader(header);

    local i = head;
    local ids = {};
    local idStr = struct.pack('I4', 0);
    local id = 0;

    while i ~= tail do
        idStr = redis.call('getrange', name, i + 8, i + 8 + 4);
        id = struct.unpack('I4', idStr);
        table.insert(ids, id);
        i = (i + 4) % len;
    end

    local msgs = {};
    local msg = '';
    for _, id in ipairs(ids) do
        msg = redis.call('get', 'msg:' .. id);

        local x = cjson.decode(msg);
        x['msgID'] = id;
        local re = cjson.encode(x);

        table.insert(msgs, re);
    end

    return msgs;
end

local function setFinish(name, msgID)
    local lastID, head, tail = 0, 0, 0;
    local header = getHeader(name);
    if header == '' then
        return nil;
    end
    lastID, head, tail = parseHeader(header);

    local idStr = redis.call('getrange', name, head + 8, head + 8 + 4);
    local id = struct.unpack('I4', idStr);

    local mID = tonumber(msgID);
    if id ~= mID then
        return nil;
    end

    local rhead = head + 4 + 8;
    local head = struct.pack('H', head + 4);
    redis.call('setrange', name, 4, head);

    return rhead;
end

if ARGV[1] == 'get_redenvelope' then
    return getRedenvelope(KEYS[1], ARGV[2]);
elseif ARGV[1] == 'push_message' then
    return pushMsg(KEYS[1], ARGV[2], ARGV[3]);
elseif ARGV[1] == 'get_message' then
    return getMsgs(KEYS[1]);
elseif ARGV[1] == 'set_finish' then
    return setFinish(KEYS[1], ARGV[2]);
else
    return redis.error_reply('unknown command');
end