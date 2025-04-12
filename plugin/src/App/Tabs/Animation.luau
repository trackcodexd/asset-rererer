--!strict
local AssetIdfilter = require("../AssetIdFilter")
local getFilterOptions = require("../GetFilterOptions")
local reuploadIds = require("../ReuploadIds")
local UiLibrary = require("../../UiLibrary")

local Selection = game:GetService("Selection")

return function(ui: UiLibrary.Ui, plugin: Plugin)
	local tab = ui:CreateTab("Animation")

	local function getFilter(instances: { Instance }): AssetIdfilter.FilterOptions
		local filterOptions = getFilterOptions(plugin, instances)
		table.insert(filterOptions.WhitelistedInstances, "Animation")
		return filterOptions
	end
	
	tab:CreateButton("Reupload", function()
		reuploadIds(plugin, ui, getFilter(game:GetDescendants()), "Animation")
	end)
	
	tab:CreateButton("Reupload Selected", function()
		reuploadIds(plugin, ui, getFilter(Selection:Get()), "Animation")
	end)
end
