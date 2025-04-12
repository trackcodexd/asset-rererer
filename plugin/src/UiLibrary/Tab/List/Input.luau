--!strict
local Input = {}

local Theme = require("../../Theme")
local Tip = require("../../Tip")

local asset = script.Parent.Parent.Parent.Assets.TextInput

local function AddTip(self: Input, tip: Tip.Tip)
    local textBox = self.instance.TextBox
    textBox.Position = UDim2.new(0, 25, 0.5, 0)
    textBox.Size = UDim2.new(1, -45, 1, -2)

    local tipFrame = tip.instance
    tipFrame.Position = UDim2.new(0, 5, 0.5, 0)
    tipFrame.Parent = self.instance

    self._tip = tip
end

local function UpdateTheme(self: Input)
    local theme = Theme.get()

    local button = self.instance
    button.BackgroundColor3 = theme.UnselectedColor
    button.BorderColor3 = theme.BorderColor

    Theme.updateText(button.TextBox, true)

    if self._tip then
        self._tip:UpdateTheme()
    end
end

local function Destroy(self: Input)
    if self._tip then self._tip:Destroy() end
    self.instance:Destroy()

    table.clear(self :: any)
end

export type Input = {
    _tip: Tip.Tip,

    instance: typeof(asset),

    AddTip: typeof(AddTip),
    UpdateTheme: typeof(UpdateTheme),
    Destroy: typeof(Destroy),
}

function Input.new(parent: GuiObject, placeholderText: string, callback: (input: string) -> ()): Input
    local instance = asset:Clone()
    local inputBox = instance.TextBox
    inputBox.PlaceholderText = placeholderText

    inputBox.FocusLost:Connect(function(enterPressed)
        if not enterPressed then return end

        local text = inputBox.Text
        if text == "" then return end
        
        inputBox.Text = ""
        callback(text)
    end)

    local self = {
        instance = instance,

        AddTip = AddTip,
        UpdateTheme = UpdateTheme,
        Destroy = Destroy,
    } :: Input

    self:UpdateTheme()
    instance.Parent = parent

    return self
end

return Input