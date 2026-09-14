# Dashboards and access

## Create your first dashboard

Sign in, create a dashboard, and choose a unique URL slug. The slug becomes part of its address, such as `/infrastructure`. Add groups to organize related services, then add links inside each group.

Use edit mode to drag groups into a new order or move links within and between groups. Layout changes are saved to the server.

## Visibility

Choose visibility based on who should reach the dashboard:

| Visibility          | Anonymous visitor | Signed-in user | Owner or admin |
| ------------------- | ----------------: | -------------: | -------------: |
| Public              |                 ✓ |              ✓ |              ✓ |
| Authenticated users |                 — |              ✓ |              ✓ |
| Private             |                 — |              — |              ✓ |

Administrators can select a non-private dashboard as the system main dashboard. This is the landing view for users who do not have a personal default and for anonymous visitors when it is public.

An individual user can mark one of their own dashboards as their personal default.

## Layout and appearance

- **Rows** places groups in horizontal sections.
- **Columns** organizes groups into columns.
- **Masonry** packs groups based on their rendered height.
- **Wide mode** uses more of the browser width.
- **Clean mode** reduces surrounding navigation for a focused display.

## Keyboard control

### Quick filter

Start typing anywhere on a dashboard to filter its items. In the regular layout, the text appears in the filter field in the navigation bar. In clean mode, DashLit shows a floating filter field while a query is active. Press **Esc** to clear the filter.

### Move between items

Use the **arrow keys** to select an item according to its position on the dashboard. Press **Enter** to open the selected item. Press **Esc** to clear the selection; it also leaves edit mode when no dialog or menu is open.

### Item and dashboard shortcuts

When creating or editing an item or dashboard on a desktop-sized screen, focus its **Hotkey** field and press a combination that includes **Ctrl**, **Alt**, **Shift**, or **Meta** plus another key. Use Backspace, Delete, or the Clear button to remove it.

An item shortcut opens every matching item on the current dashboard. A dashboard shortcut switches to one of your own dashboards. Dashboard shortcuts are unique for each user and are never shown or handled for another user's dashboards or for anonymous visitors. A dashboard shortcut cannot reuse a shortcut assigned to one of your items.

## Themes and custom themes

Open the theme menu from the navigation bar or the floating control in clean mode to select one of the built-in light or dark themes. Choose **Edit colors…** to create a custom theme. If no custom theme exists yet, the editor starts with the colors of the currently selected theme.

The editor lets you adjust the page, surface, text, primary, accent, danger, and success colors, select a light or dark color scheme, and optionally add a background image. A background can be referenced by URL. Signed-in users can also upload an image to DashLit; file upload is unavailable to anonymous visitors.

After saving, **Custom** appears as a regular option in the theme menu and can be selected without reopening the editor. Use **Delete theme** in the editor to remove it and return to the system theme. For signed-in users the selected theme, custom palette, and background reference are stored with their account and follow them between browsers. Anonymous preferences are stored only in the current browser.

## Icon search and theme variants

The icon picker searches two sources independently:

- **selfh.st/icons** focuses on software and services commonly used in self-hosted environments.
- **Iconify** provides a much broader collection of general-purpose icon sets.

Iconify results are displayed as soon as they arrive. The selfh.st search continues independently and adds its results without blocking or replacing Iconify results.

For selfh.st icons, DashLit prefers SVG and falls back to PNG when SVG is unavailable. Selecting a selfh.st icon for the Default field automatically fills the empty Dark field with the corresponding theme variant. An existing manually selected Dark icon is never overwritten.

The selfh.st convention uses a dark-colored asset on light backgrounds and a light-colored asset on dark backgrounds. When no separate theme asset exists, DashLit uses the standard icon for both fields. Downloaded selfh.st assets are served through the DashLit API and cached on disk.

Many monochrome Iconify assets are black by default. DashLit renders Iconify icons in a light color on dark application themes so they remain visible. This adjustment applies only to Iconify; branded selfh.st icons retain their supplied colors and theme variants.

The picker shows Default icons on a light preview background and Dark icons on a dark preview background. You can still override either field manually, upload a file, or enter a remote image URL.

## Availability checks

Enable the availability check on a link to display whether its destination is reachable. By default DashLit checks the link URL, but you can provide a separate check URL—for example, an internal address or a dedicated health endpoint. Enable **Skip TLS verification** only when that endpoint uses a trusted self-signed or mismatched certificate. “Only when down” keeps healthy links quiet and surfaces the indicator only when a check fails.

Availability is a convenience signal from the DashLit server, not a replacement for full monitoring and alerting.

## Import, export, and cloning

Dashboard export creates a portable DashLit JSON file containing dashboard settings, groups, and items. Import it while signed in to create a new owned dashboard. Slug conflicts are resolved automatically.

You can also clone complete dashboards, individual groups, or individual items when building similar views. From a group's actions menu, choose **Copy to dashboard…** to select another dashboard you can edit and copy the group there together with all of its links and settings.
