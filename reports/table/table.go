package table

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func ServeCSVFile(before, after string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print("looking for csv file")
		if strings.HasSuffix(r.URL.Path, ".csv") {
			file, err := os.Open("page" + strings.Replace(r.URL.Path, ".html", ".csv", 1))
			if err != nil {
				http.Error(w, "File not found", http.StatusInternalServerError)
				return
			}
			io.Copy(w, file)
			return
		}
		file, err := os.Open("page" + strings.Replace(r.URL.Path, ".html", ".csv", 1))
		if err != nil {
			http.Error(w, "File not found", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		writter := bufio.NewWriter(w)
		defer writter.Flush()

		var line []string
		filebefore, err := os.Open(before)
		if err != nil {
			http.Error(w, "Before file doesn't exist", http.StatusInternalServerError)
			return
		}
		writter.ReadFrom(filebefore)

		line, err = reader.Read()

		writter.WriteString(`
		<script>
    var table
    $(document).ready( function () {
        table = $('#table_id').DataTable({
            fixedHeader: {
                header: true,
                footer: false
            },
            colReorder: {
                enable: true
            },
            dom: 'Bfrtip',

            buttons: [
                {
                    extend: 'excelHtml5',
                    filename: 'sample',
                    text: 'Save as Excel'
                }
            ], 
			columns:[
			`)

		for i, record := range line {
			if i < len(line)-1 {
				writter.WriteString("{name:'" + record + "'},")
			} else {
				writter.WriteString("{name:'" + record + "'}")
			}
		}

		writter.WriteString(`
	]
            });
    } );
	</script>
		`)

		writter.WriteString(`<table id="table_id" class="display">`)
		writter.WriteString("\n<thead>\n")
		writter.WriteString("<tr>\n")
		for _, record := range line {
			writter.WriteString("<th>")
			writter.WriteString(record)
			writter.WriteString("</th>\n")
		}
		writter.WriteString("</tr>\n")
		writter.WriteString("</thead>\n")

		writter.WriteString("<tbody>\n")
		lines := 0
		for err == nil {
			line, err = reader.Read()
			lines++
			if len(line) > 0 {
				writter.WriteString("<tr>\n")
				for _, record := range line {
					writter.WriteString("<td>")
					writter.WriteString(record)
					writter.WriteString("</td>\n")
				}
				writter.WriteString("</tr>\n")
			}
		}
		log.Print(err)
		log.Printf("lines %d, last line %s", lines, line)
		writter.WriteString("</tbody>\n")
		writter.WriteString("</table>\n")

		fileafter, err := os.Open(after)
		if err != nil {
			http.Error(w, "After file doesn't exist", http.StatusInternalServerError)
			return
		}
		writter.ReadFrom(fileafter)

	})
}
