package subteez

type NotFoundError struct{}

func (*NotFoundError) Error() string {
	return "Requested resource not found"
}
