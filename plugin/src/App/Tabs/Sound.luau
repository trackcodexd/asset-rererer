--!strict
local AssetIdfilter = require("../AssetIdFilter")
local getFilterOptions = require("../GetFilterOptions")
local reuploadIds = require("../ReuploadIds")
local UiLibrary = require("../../UiLibrary")

local MarketplaceService = game:GetService("MarketplaceService")
local Selection = game:GetService("Selection")

return function(ui: UiLibrary.Ui, plugin: Plugin)
    local tab = ui:CreateTab("Sound")
    local placeList

    local function getFilter(instances: { Instance }): AssetIdfilter.FilterOptions
        local filterOptions = getFilterOptions(plugin, instances)
        table.insert(filterOptions.WhitelistedInstances, "Sound")
        return filterOptions
    end

    local function getPlaceList(): { number }
        local placesListValues = placeList:Get()
        local list = table.create(#placesListValues)
        for _, value in placesListValues do
            table.insert(list, tonumber(value))
        end
        return list :: { number }
    end

    tab:CreateButton("Reupload", function()
		reuploadIds(plugin, ui, getFilter(game:GetDescendants()), "Sound", getPlaceList())
	end)
	
	tab:CreateButton("Reupload Selected", function()
        reuploadIds(plugin, ui, getFilter(Selection:Get()), "Sound", getPlaceList())
	end)

    placeList = tab:CreateList()

	local placeInput = placeList:AddInput("Place Id", function(input)
		local placeId = tonumber(input)
        local success, productInfo = pcall(MarketplaceService.GetProductInfo, MarketplaceService, placeId)

        if not success then
            ui:Notify("Notification", "Error getting product info.") 
            return 
        end

        local assetTypeId = (productInfo :: any).AssetTypeId
		if assetTypeId ~= 9 then
            ui:Notify("Notification", "Invalid place id.") 
            return 
        end

        placeId = tostring(placeId)
        if placeList.elements[placeId] then
            ui:Notify("Notification", "Place id is in list.") 
            return 
        end

        local element = placeList:CreateElement(placeId)
            :AddDestroyButton()
        
        element.instance.TextLabel.Text = `({placeId}) {(productInfo :: any).Name}`
	end)
	placeInput:AddTip(ui:CreateTip(
		"Reuploading audio needs a place id. Public games are found automatically."
	))

    local textBox = placeInput.instance.TextBox
    textBox:GetPropertyChangedSignal("Text"):Connect(function() 
        textBox.Text = string.gsub(textBox.Text, "%D", "") 
    end)
end
