Status: Somewhat fixed?

Old State:
> so it builds after the refactor but the start game button doesn't reload players' screens anymore to watch the countdown. when i refresh the browsers, they are indeed in the game with roles assigned after 5 seconds.
> 
> it also doesn't refresh as players join/leave the game.
> 
> it also doesn't update the counts or any other settings when you adjust them in the lobby settings!
> 
> i fear we've totally broken some of the SSE behavior?! are we testing for this? did you check?
> The server seems to process the events, and players refreshing does cause them to see the new info after a reload, but SSE is totally broken now.

I did some changes and you can examine the current repo state, but I really need you to examine the SDK here: https://github.com/starfederation/datastar-go/


As well as all the release notes for the RC releases of Datastar, as there were some changes that I think we need/needed to make: 
https://github.com/starfederation/datastar/releases

RC changes to look out for
================================================================================================================================
v1.0.0-RC.3

This release packs some new features, cleans up data-bind behavior, and fixes a rake of minor bugs.

    Added support for patching html, head, and body elements.
    Added support for patching arbitrary fragments into the DOM given a selector is provided.
    Added support for passing include and exclude filters in as strings.
    Added a __filter modifier to data-query-string that filters out empty values when syncing signals to query string params.
    When a signal is predefined, its type is now preserved during binding – whenever the element’s value changes, the signal value is automatically converted to match the original type.
    Custom keys in data-persist are now added as data-persist-mykey, and the __key modifier has been removed.
    Numeric signal names are now correctly parsed.
    Fixed a regression in which data-attr values were not being JSON encoded.
    Fixed a bug in which data-text and data-json-signals would not reapply themselves after their text was changed.
    Fixed a bug in which attributes that had a prefix with the same length as the alias would not be ignored (e.g. data-option-load would be applied as on-load).
    Fixed a bug in which the retries-failed event was named retrying.
    Fixed a bug in which the remove mode would fail when patching elements if no elements were provided.
    Fixed a bug in which data-class would enter an infinite loop if there were multiple such attributes on the same element.
    Fixed a bug in which attribute plugins on child elements were not cleaned up when removed.
    Removed support for using the ID of a provided element as a selector for the remove mode when patching elements.
    Changed the behavior of how input values are diffed when morphed:

Old input value 	New input value 	Behavior
null 	null 	preserve old input value
some value 	the same value 	preserve old input value
some value 	null 	set old input value to ""
null 	some value 	set old input value to new input value
some value 	some other value 	set old input value to new input value




v1.0.0-RC.2

    Added a data-style attribute that sets the value of inline CSS styles on an element based on an expression, and keeps them in sync.
    Added a requestCancellation option to backend actions that controls request cancellation behavior.
    Fixed a bug in which indicator signals were not being set to false after a text/html response completed.
    Fixed a bug in which a duplicate text/event-stream entry was being added to the Accept header in SSE actions.
    Fixed a bug in which assigning a computed to an already existing signal would create a signal of a computed instead of a computed.
    Fixed a bug in which using data-attr with a key would observe the result of the expression rather than the key.
    Fixed a bug in which signal names were being incorrectly parsed in some edge-cases.
    Renamed the datastar-sse event to datastar-fetch.


v1.0.0-RC.1

Some plugins are now available under Datastar Pro, which adds functionality to the the free open source Datastar framework. These plugins are available under a commercial license that helps fund our open source work.

Of the many changes listed below, one major feature is that objects in signals are now reactive! This means that you can now create complex data structures in signals, and any changes to these objects will automatically propogate to expressions.

SSE event handling has also changed, in addition to all of the SDKs. Please refer to the SSE docs and each of the SDKs for the correct syntax to use.

    Objects in signals are now reactive, meaning that any changes to these objects will automatically propogate to expressions.
    Plugins are now reapplied on morph only if their values/keys/modifiers have changed.
    Added the ability for Datastar to receive text/html, application/json, and text/javascript content types, that patch elements, patch signals, and execute JavaScript respectively.
    Added a data-effect attribute that executes an expression when any of the signals it references change.
    Added a data-ignore-morph attribute to the PatchElements watcher that skips morphing the respective element and its children.
    Added a data-json-signals attribute that sets the text content of an element to a reactive JSON stringified version of all signals.
    Added a data-on-signal-patch attribute that executes an expression when a signal patch takes place.
    Added a data-on-signal-patch-filter attribute for filtering the signals that cause the expression in data-on-signal-patch to be executed.
    Added a data-preserve-attr attribute that preserves the client side state of an attribute through a morph.
    Added a data-on-resize attribute (PRO) that attaches a ResizeObserver to the element, and executes the expression each time the element’s dimensions change.
    Added a data-query-string attribute (PRO) that syncs the query string with signal values, including optional history support.
    Added a datastar-upload-progress event (PRO) for monitoring file upload progress.
    Added a filterSignals option to SSE actions that filters the signals send to the backend based on include and exclude regular expressions.
    Added a __trusted modifier to the data-on attribute, which runs the expression even if the isTrusted property on the event is false.
    Added automatic request cancellation for SSE actions - when a new request is initiated on an element, any existing request on that same element is automatically cancelled.
    Removed the abort option from SSE actions as request cancellation is now handled automatically at the element level.
    The URL passed into SSE actions (@get, @post, etc.) is now treated as a relative URI.
    The default Content-Type header sent with form requests is now application/x-www-form-urlencoded.
    The value of a clicked button element is now included in the request when using the form content type.
    The data-star-ignore attribute has been renamed data-ignore.
    The data-attr attribute now renders true as "" instead of "true" (e.g. checked="" instead of checked="true").
    The data-attr attribute now preserves the string literals "false", "null", and "undefined" when using a key.
    Fixed a bug when using the __debounce.leading modifier with the data-on attribute.
    Removed the data-on-signal-change attribute. Use the new data-on-signal-patch attribute instead.
    Removed the datastar-signal-change event. Use the new datastar-signal-patch event instead.
    Removed the includeLocal option in backend action requests. Use the filterSignals option instead.
    Removed the variable ctx from data attributes. Use the new el variable to access the element the attribute is attached to, use the new $ variable to access the signal root, or the new data-json-signals attribute to output all signals.
    Removed support for adding a dollar sign prefix to signal names in the value of the data-bind, data-ref, and data-indicator attributes.
    Removed the auto generated IDs that were assigned to elements using data attributes.

Changes to SSE Event Handling

    Renamed the datastar-merge-fragments and datastar-merge-signals SSE events to datastar-patch-elements and datastar-patch-signals respectively.
    Renamed the mergeMode parameter of the datastar-patch-elements SSE event to mode.
    Renamed the morph mode to outer.
    Renamed the outer mode to replace.
    The inner mode now morphs the element’s inner HTML.
    Removed the upsertAttributes mode.
    Added the remove mode.
    The datastar-patch-signals SSE event now patches (adds/updates/removes) signals according to the JSON Merge Patch RFC 7396.
    Removed the RemoveFragments, RemoveSignals, and ExecuteScript watchers.


Timeouts
=
While clicking around I get a LOT of:
Datastar failed to reach http://localhost:7331/sse/lobby/JZI3D?datastar=%7B%22theme%22%3A%22night%22%2C%22isStarting%22%3Afalse%2C%22canStartGame%22%3Atrue%2C%22canAutoScale%22%3Afalse%2C%22cardId%22%3A%22%22%2C%22cardChecked%22%3Afalse%2C%22roleType%22%3A%22%22%2C%22roleCount%22%3A0%2C%22action%22%3A%22%22%2C%22accordionLeader%22%3Afalse%2C%22accordionGuardian%22%3Afalse%2C%22accordionAssassin%22%3Afalse%2C%22accordionTraitor%22%3Afalse%2C%22allowLeaderless%22%3Afalse%2C%22hideRoleDistribution%22%3Afalse%2C%22fullyRandomRoles%22%3Afalse%2C%22updatingLeaderless%22%3Afalse%2C%22updatingHideDistribution%22%3Afalse%2C%22updatingFullyRandom%22%3Afalse%2C%22startError%22%3A%22%22%2C%22validationMessage%22%3A%22%22%7D retrying in 1000ms.



In the browser consoles--are we hitting some kind of timeout?

I tested this by changing some stuff in 
nix/app/internal/config/viper_config.go

As such:

	// Timeout defaults
	v.SetDefault("server.readtimeout", "0s")
	v.SetDefault("server.writetimeout", "0s")
	v.SetDefault("server.idletimeout", "0s") // 0 for SSE support
	v.SetDefault("server.shutdowntimeout", "0s")


But I'm not sure which are necessary and best practices--you'll need to do some research on Datastar, SSE, and Viper to figure the best way forward out, I think.


My changes seemingly got the auto-updating on join/leave/game start working, but I think something is still amiss because theme switching also does not work.

Maybe a data-bind issue?

In order to get the browser to not throw errors, I had to rename the theme binding from

$theme

to just

theme


IDK!
https://data-star.dev/reference/attributes#data-bind

It feels like there's something wonky with the data binding I guess? Not entirely sure, but it broke seemingly randomly and I'm not even sure hot it was working in the first place so please take a look.

DO NOT revert Datastar to an older version or change the versions to fix this, and DO NOT simply strip functionality to fix this.


