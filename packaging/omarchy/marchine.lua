-- Optional Omarchy/Hyprland window rule for Marchine.
-- Copy to ~/.config/hypr/marchine.lua and load it from your Hyprland Lua
-- configuration (for example: require("marchine")).
--
-- Marchine's desktop launcher requests true fullscreen from Foot. This optional
-- rule gives the floating backing window the current monitor geometry as well,
-- so a compositor fullscreen handoff cannot reveal an undersized launcher.
--
-- Users can still leave fullscreen and resize Marchine normally afterward.
--
-- This is an example only. Marchine's package must not modify a user's
-- compositor configuration.

hl.window_rule({
    name = "marchine-launcher",

    match = {
        initial_class = "^marchine$",
    },

    float = true,
    size = { "monitor_w", "monitor_h" },
    move = { 0, 0 },
})
