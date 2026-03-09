package security

func NamespaceIsolate(runtimeId string, operation string) (IsolatedView, error) {
	return IsolatedView{
		RuntimeID: runtimeId,
		Operation: operation,
		Boundary:  DMZBoundary,
	}, nil
}

func (iv *IsolatedView) FilterData(data any) any {
	return nil
}

func (iv *IsolatedView) GetOperation() string {
	return ""
}
