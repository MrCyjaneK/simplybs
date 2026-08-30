package crash

import "log"

func Handle(err error) {
	if err != nil {
		log.Panicln(err)
	}
}
