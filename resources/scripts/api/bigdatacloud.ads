-- Copyright © by Jeff Foley 2017-2023. All rights reserved.
-- Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.
-- SPDX-License-Identifier: Apache-2.0

local json = require("json")
local url = require("url")

name = "BigDataCloud"
type = "api"

function start()
    set_rate_limit(1)
end

function check()
    local c
    local cfg = datasrc_config()
    if (cfg ~= nil) then
        c = cfg.credentials
    end

    if (c ~= nil and c.key ~= nil and c.key ~= "") then
        return true
    end
    return false
end

function asn(ctx, addr, asn)
    local c
    local cfg = datasrc_config()
    if (cfg ~= nil) then
        c = cfg.credentials
    end

    if (c == nil or c.key == nil or c.key == "") then
        return
    end

    local resp, err = request(ctx, {['url']=build_url(addr, c.key)})
    if (err ~= nil and err ~= "") then
        log(ctx, "asn request to service failed: " .. err)
        return
    elseif (resp.status_code < 200 or resp.status_code >= 400) then
        log(ctx, "as request to service returned with status: " .. resp.status)
        return
    end

    local d = json.decode(resp.body)
    if (d == nil) then
        log(ctx, "failed to decode the JSON response")
        return
    elseif (d.carriers == nil or #(d.carriers) == 0) then
        return
    elseif (d.registry == nil or d.bgpPrefix == nil or 
        d.registeredCountry == nil or d.organisation == nil) then
        return
    end

    new_asn(ctx, {
        ['addr']=addr,
        ['asn']=d['carriers'][1].asnNumeric,
        ['cc']=d.registeredCountry,
        ['desc']=d.organisation,
        ['registry']=d.registry,
        ['prefix']=d.bgpPrefix,
    })
end

function build_url(addr, key)
    local params = {
        ['localityLanguage']="en",
        ['ip']=addr,
        ['key']=key,
    }

    return "https://api.bigdatacloud.net/data/network-by-ip?" .. url.build_query_string(params)
end
