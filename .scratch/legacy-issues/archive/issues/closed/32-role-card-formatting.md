Text portion of the formatting on the player role card page is a bit jumbled. Compare the screenshots:

issues/closed/Screenshot_20250715_152555.png
issues/closed/Screenshot_20250715_152720.png

## Resolution
Fixed in commit b2831df9 - ✨ feat: improve role card text formatting

### Changes made:
1. Added Mana font for proper mana symbol rendering ({2}, {W}, etc.)
2. Formatted card text with line breaks at pipe characters (|)
3. Italicized reminder text in parentheses
4. Fixed Unicode character rendering (bullets, em dash, quotes)
5. Added inline CSS fallbacks for when external CSS fails to load
6. Used templ.Raw with safe HTML escaping for Unicode preservation

The issue was caused by templ's HTML escaping converting Unicode characters to HTML entities. Fixed by using @templ.Raw() with a custom escaping function that preserves Unicode while still escaping dangerous HTML.
