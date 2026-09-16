-- Optional Omarchy/Hyprland window rule for Marchine.
-- Copy to ~/.config/hypr/marchine.lua and load it from your Hyprland Lua
-- configuration (for example: require("marchine")).
--
-- This is an example only. Marchine's AUR package must not modify a user's
-- compositor configuration.

hl.window_rule({
    name = "marchine-launcher",

    match = {
        initial_class = "^marchine$",
    },

    float = true,
    size = { 1180, 900 },
    center = true,
})
