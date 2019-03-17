$(document).ready(function(){
     table = $('#frdtable').DataTable({
        columns: [
            {"name": "name"},
            {"name": "tf"}
        ],
        "language":{
            "emptyTable": 'Add friends to split amount',
        },
        "bInfo": false,
        "bFilter": true,
        "paging":   false,
        "info":     false,
        "searching": false
    });
    $('#message').hide();
    $('#output').hide();
    table.rows().remove().draw();
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
            var sptype = $('input[type=radio][name=split]:checked').val();
            var amt = $('#amount').val();
            if (sptype == "equal") {
                table.row.add([
                    selected,
                    '<input type="text" name="'+selected+'" id="'+selected+'">'
                ]).draw();
            } else {
                table.row.add([
                    selected,
                    '<input type="text" name="'+selected+'" id="'+selected+'" value="0">'
                ]).draw();
            }
        } else {
            table.row(rowno).remove().draw();
        }
        updateTable();
    })
    $('input[type=radio][name=split]').change(function(){
        var sptype = $('input[type=radio][name=split]:checked').val();
        var amt = $('#amount').val();
        updateTable();
    });

    $('#amount').change(function(){
        updateTable();
    });

    $('#date').datepicker({
        uiLibrary: 'bootstrap4',
        format: 'dd/mm/yyyy',
    });
    
    $("#date").prop("readonly", true);
    $('#submit').click(function(){
        var desc = $('#desc').val();
        var totamt = $('#amount').val();
        var splitType = $('#split').val();
        var bdate = $('#date').val();
        var names = [];
        var splitval = [];
        var row_count = table.rows().count();
        if (desc != "" && totamt != "" && splitType != "" && bdate != "" && !isNaN(totamt)) {
            for(i = 0; i < row_count; i++) {
                var temp = (table.row(i).data()[0]);
                splitval.push($('#'+temp).val());
                names.push(temp);
            }
            $.ajax({
                url: 'uploadbill',
                type: 'post',
                dataType: 'html',
                data: {des: desc, tamt: totamt, billdt: bdate, friends: names, amtsplit: splitval},
                success: function(data) {
                    $('#message').show();
                    $('#output').show();
                    $('#message').text(data);
                },
                error: function(error){
                    alert(error);
                }
            });
        }
    });
});

function updateTable(){
    var sptype = $('input[type=radio][name=split]:checked').val();
    var row_count = table.rows().count();
    var amt = $('#amount').val();
    var names = []
    divamt = amt/row_count;
    if (!isNaN(divamt)) {
        for (i = 0; i < row_count; i++) {
            names.push(table.row(i).data()[0]);
        }
        if (sptype == "equal") {
            for (i = 0; i < row_count; i++) {
                $('#'+names[i]).val(divamt);
            }
        } else if (sptype == "unequal") {
            for (i = 0; i < row_count; i++) {
                $('#'+names[i]).val('0');
            }
        }
    }
    
}