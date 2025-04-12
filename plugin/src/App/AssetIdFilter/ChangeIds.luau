--!strict
local ApiDump = require("../ApiDump")
local retry = require("../Retry")

local ScriptEditorService = game:GetService("ScriptEditorService")
local AssetService = game:GetService("AssetService")

local cachedClassProperties: { [string]: { string } } = {}

export type IdPair = {
    newId: number,
    oldId: number
}

local function setScriptIds(scriptInstance: Script | LocalScript | ModuleScript, idsToChange: { IdPair })
    local source = scriptInstance.Source
    if not source then
        warn(scriptInstance, "has nil source.")
        return
    end

    for _, idPair in idsToChange do
        local oldId = idPair.oldId
        local newId = idPair.newId
        source = string.gsub(source, `%f[%d]{oldId}%f[%D]`, tostring(newId))
    end

    if #source > 200_000 then
        ScriptEditorService:UpdateSourceAsync(scriptInstance, function() return source end) -- we do a bulk update because this function very very slow :/
    else
        scriptInstance.Source = source
    end
end

local function setAnimationId(animation: Animation, oldId: number, newId: number)
    animation.AnimationId = string.gsub(animation.AnimationId, tostring(oldId), tostring(newId)) 
end

local function setSoundId(sound: Sound, oldId: number, newId: number)
    sound.SoundId = string.gsub(sound.SoundId, tostring(oldId), tostring(newId))  
end

local function setNumberValueId(numValue: NumberValue | IntValue, _: number, newId: number)
    numValue.Value = newId
end

local function setStringValueIds(strValue: StringValue, oldId: number, newId: number)
    strValue.Value = string.gsub(strValue.Value, tostring(oldId), tostring(newId)) 
end

local function setCharacterMesh(characterMesh: CharacterMesh, _: number, newId: number)
    characterMesh.MeshId = newId
end

local function transferAttributes(oldInstance: Instance, newInstance: Instance)
    for name, value in oldInstance:GetAttributes() do
        newInstance:SetAttribute(name, value)
    end
end

local function transferTags(oldInstance: Instance, newInstance: Instance)
    for _, tag in oldInstance:GetTags() do
        newInstance:AddTag(tag)
    end
end

local function transferChildren(oldInstance: Instance, newInstance: Instance)
    for _, child in oldInstance:GetChildren() do
        child.Parent = newInstance
    end
end

local function transferProperties(oldInstance: Instance, newInstance: Instance)
    local className = oldInstance.ClassName
    if className ~= newInstance.ClassName then error(`oldInstance({className}) class is not equal to newInstance({newInstance.ClassName})`) end
    assert(ApiDump.isCached(), "no api dump cached")

    local cachedProperties = cachedClassProperties[className]
    if not cachedProperties then
        cachedProperties = ApiDump.getProperties(className)
        cachedClassProperties[className] = cachedProperties
    end

    for _, property in cachedProperties do
        if property == "Parent" then continue end
        (newInstance :: any)[property] = (oldInstance :: any)[property]
    end
end

local function setMeshPart(meshPart: MeshPart, oldId: number, newId: number)
    local contentRetrieved, content: Content = retry(3, Content.fromAssetId, newId) 
	if not contentRetrieved then
		warn(`failed to get content from {newId}, skipping {oldId}`)
        return
	end
    
    local newMeshPart = AssetService:CreateMeshPartAsync(content, {
        CollisionFidelity = meshPart.CollisionFidelity,
        RenderFidelity = meshPart.RenderFidelity,
        FluidFidelity = meshPart.FluidFidelity
    } :: any)

    local propertiesTransferred, result = pcall(transferProperties, meshPart, newMeshPart)
    if not propertiesTransferred then
        newMeshPart:Destroy()
        error(result)
    end

    transferAttributes(meshPart, newMeshPart)
    transferTags(meshPart, newMeshPart)
    transferChildren(meshPart, newMeshPart)

    newMeshPart.Parent = meshPart.Parent
    meshPart:Destroy()
end

local function setSpecialMesh(specialMesh: SpecialMesh, oldId: number, newId: number)
    specialMesh.MeshId = string.gsub(specialMesh.MeshId, tostring(oldId), tostring(newId))
end

local instanceIdSetters = {
    Animation = setAnimationId,
    Sound = setSoundId,

    Script = setScriptIds,
    LocalScript = setScriptIds,
    ModuleScript = setScriptIds,

    NumberValue = setNumberValueId,
    IntValue = setNumberValueId,
    StringValue = setStringValueIds,

    CharacterMesh = setCharacterMesh,
    MeshPart = setMeshPart,
    SpecialMesh = setSpecialMesh,
}

return function(filteredIds: { [number]: { Instance } }, idsToChange: { IdPair })
    local scriptIdChanges = {}

    for _, idPair in idsToChange do
        local oldId = idPair.oldId
        local newId = idPair.newId

        local instanceArray = filteredIds[oldId]
        if not instanceArray then continue end

        for _, instance in instanceArray do
            if instance:IsA("BaseScript") or instance:IsA("ModuleScript") then
                if not scriptIdChanges[instance] then scriptIdChanges[instance] = {} end
                table.insert(scriptIdChanges[instance], idPair)
                continue
            end

            local className = instance.ClassName
            local setInstanceId = instanceIdSetters[className]
            if not setInstanceId then error(`{className} is not supported`) end
            setInstanceId(instance, oldId, newId)
        end
    end

    for instance, idPairs in scriptIdChanges do
        setScriptIds(instance, idPairs)
    end
end
