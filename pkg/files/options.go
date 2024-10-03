package files

type ListFilesOptions struct {
	WithContent bool
	ShowHidden  bool
	Filter      PatternFilter
}

type ListFileOption func(opts *ListFilesOptions)

func ListFilesWithContent() ListFileOption {
	return func(opts *ListFilesOptions) {
		opts.WithContent = true
	}
}

func ListFilesWithShowHidden() ListFileOption {
	return func(opts *ListFilesOptions) {
		opts.ShowHidden = true
	}
}

func ListFilesWithFilter(filter PatternFilter) ListFileOption {
	return func(opts *ListFilesOptions) {
		opts.Filter.Include = append(opts.Filter.Include, filter.Include...)
		opts.Filter.Exclude = append(opts.Filter.Exclude, filter.Exclude...)
	}
}
