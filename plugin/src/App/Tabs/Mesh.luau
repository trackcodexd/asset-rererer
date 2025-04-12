--!strict
local AssetIdfilter = require("../AssetIdFilter")
local getFilterOptions = require("../GetFilterOptions")
local reuploadIds = require("../ReuploadIds")
local UiLibrary = require("../../UiLibrary")
local ApiDump = require("../ApiDump")

local Selection = game:GetService("Selection")

local meshSettings: { [string]: string } = {
    ["MeshParts"] = "MeshPart",
    ["SpecialMeshes"] = "SpecialMesh",
    ["CharacterMeshes"] = "CharacterMesh"
}

return function(ui: UiLibrary.Ui, plugin: Plugin)
	local tab = ui:CreateTab("Mesh")

    if not ApiDump.get() then
        local thread = coroutine.running()
        local button; button = tab:CreateButton("Failed to get api dump: retry", function()
            if not ApiDump.get() then return end
            (button :: any):Destroy()
            coroutine.resume(thread)
        end)
        coroutine.yield(thread)
    end

    local split = tab:CreateSplit()
    split:CreateToggle("Search mesh parts", function(state)
        plugin:SetSetting("MeshParts", state)
    end, plugin:GetSetting("MeshParts"))
    split:CreateToggle("Search special meshes", function(state)
        plugin:SetSetting("SpecialMeshes", state)
    end, plugin:GetSetting("SpecialMeshes"))

    local split2 = tab:CreateSplit()
    split2:CreateToggle("Search character meshes", function(state)
        plugin:SetSetting("CharacterMeshes", state)
    end, plugin:GetSetting("CharacterMeshes"))

    local function getFilter(instances: { Instance }): AssetIdfilter.FilterOptions
        local filterOptions = getFilterOptions(plugin, instances)
        for settingName, className in meshSettings do
            if not plugin:GetSetting(settingName) then continue end
            table.insert(filterOptions.WhitelistedInstances, className)
        end
        return filterOptions
    end
	
	tab:CreateButton("Reupload", function()
		reuploadIds(plugin, ui, getFilter(game:GetDescendants()), "Mesh")
	end)
	
	tab:CreateButton("Reupload Selected", function()
		reuploadIds(plugin, ui, getFilter(Selection:Get()), "Mesh")
	end)
end
