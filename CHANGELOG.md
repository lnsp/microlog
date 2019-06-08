# Changelog
All notable changes to this project will be documented in this file.

## 2019-06-09
### Added
- Protection against CSRF attacks is enabled by default

## 2019-05-08
### Changed
- Complete rework of color scheme
- New favicon design
- New contextual form components

## 2018-11-11
### Changed
- Moved session tokens into separate service

## 2018-11-01
### Added
- Moderation panel for community leaders

### Changed
- Post titles now have a minimum length of 3 characters, while post content must be at least 10 characters

## 2018-10-03
### Added
- Dashboard view now supports selection of interval for popular posts

### Changed
- Profile view now uses a more consistent visual language

## 2018-09-23
### Added
- New logo resembling typographic redesign

### Changed
- Improved typography across the site

## 2018-09-14
### Added
- Improve internal logging capabilities
- Add link to GitHub issues project page to footer
- Add Markdown notice on post edit page

## Changed
- Remember form values after bad signup (name already in use, bad email etc.)

## 2018-08-24
### Added
- 'Like' button added for expression of joy and satisfaction over a post

### Changed
- Replace 'Recent Posts' on dashboard with 'Most popular posts this week'

### Fixed
- Authors can not report their own posts anymore

## 2018-08-23
### Added
- Publicly available changelog on [microlog.co/changelog](https://microlog.co/changelog)
- Reorganize profile view and add 'Settings' subsection
- Make specific routed names unavailable for sign up like `microlog`, `auth` and `profile`
- Support on-the-fly minification of responses
