$(document).ready(function(){
    var table = $('#frdtable').DataTable({
        columns: [
            {"name": "name"},
            {"name": "tf"}
        ],
        "emptyTable": 'Add friends to split amount',
        "bInfo": false,
        "bFilter": true,
        "paging":   false,
        "info":     false,
        "searching": false
    });
    $('#add').click(function(){
        var selected = $('#friends').find(":selected").text();
        row_count = table.rows().count();
        var found = false;
        var rowno = -1;
        for (i = 0; i < row_count; i++) {
            if (table.row(i).data()[0] == selected) {
                found = true
                rowno = i
                break
            }
        }
        if (!found && selected != "Select any") {
            table.row.add([
                selected,
                '<input type="text" name="'+selected+'" id="'+selected+'">'
            ]).draw();
        } else {
            table.row(rowno).remove().draw();
        }
    })
});